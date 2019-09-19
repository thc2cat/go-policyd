package main

// History :
// 2019/09/10 :  tag 0.1 - compiling.
// 2019/09/12 :  tag 0.1 - deployed
// 2019/09/13 :  tag 0.3 - +pid,whitelist/blacklist
// 2019/09/13 :  tag 0.4 - +correction bug SUM (cast)
// 2019/09/16 :  tag 0.5 - +dbClean
// 2019/09/17 :  tag 0.6 - cut saslUsername@uvsq.fr
// 2019/09/19 :  tag 0.61 - more logs for whitelist/blacklist
//                       - auto version with git tag
//
// TODO : with context for DB blackout.

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type connData struct {
	saslUsername   string
	clientAddress  string
	sender         string
	recipientCount string
}

var (
	xdebug       *bool
	xmutex       sync.Mutex
	defaultQuota int64

	// Version is git tag version exported by Makefile, printed by -debug
	Version string
)

const (
	cfgfile = "/etc/postfix/policyd.cfg"
)

func main() {

	InitCfg(cfgfile)
	xdebug = flag.Bool("debug", false, "enable debugging")
	flag.Parse()
	defaultQuota, _ = strconv.ParseInt(cfg["defaultquota"], 0, 64)

	if !*xdebug {
		daemon(0, 0)
	} else {
		fmt.Printf("Starting %s in foreground mode\n", Version)
	}
	// Listen for incoming connections.
	l, err := net.Listen("tcp", cfg["bind"]+":"+cfg["port"])
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	initSyslog("policyd")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", cfg["dbuser"], cfg["dbpass"], cfg["dbname"]))
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	go dbClean(db)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			xlog.Err("Error accepting: " + err.Error())
			os.Exit(1)
		}
		go handleRequest(conn, db)
	}
}

func dbClean(db *sql.DB) {
	for {
		xmutex.Lock()
		db.Ping()
		// Keep 7 days in db
		db.Exec("DELETE from events where  ts<SUBDATE(CURRENT_TIMESTAMP(3), INTERVAL 7 DAY)")
		xmutex.Unlock()
		// Clean every day
		time.Sleep(24 * time.Hour)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, db *sql.DB) {
	var xdata connData

	// fmt.Println("->Entering handleRequest")
	// defer fmt.Println("<-Exiting handleRequest")

	reader := bufio.NewReader(conn)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
		s = strings.Trim(s, " \n\r")
		s = strings.ToLower(s)
		if s == "" {
			break
		}
		vv := strings.SplitN(s, "=", 2)
		if len(vv) < 2 {
			xlog.Err("Error processing line" + s)
			continue
		}
		// if (*xdebug) { fmt.Println("..", s, ":", vv[0],"->",vv[1]) }
		vv[0] = strings.Trim(vv[0], " \n\r")
		vv[1] = strings.Trim(vv[1], " \n\r")
		switch vv[0] {
		case "sasl_username":
			if strings.IndexByte(vv[1], '@') == -1 {
				xdata.saslUsername = vv[1]
			} else {
				xdata.saslUsername = vv[1][:strings.IndexByte(vv[1], '@')]
			}
		case "sender":
			xdata.sender = vv[1]
		case "client_address":
			xdata.clientAddress = vv[1]
		case "recipient_count":
			xdata.recipientCount = vv[1]
		}
	}

	resp := policyVerify(xdata, db) // Here, where the magic happen

	conn.Write([]byte(fmt.Sprintf("action=%s\n\n", resp)))
	conn.Close()
}

func policyVerify(xdata connData, db *sql.DB) string {

	var dbSum int64

	// fmt.Println("->Entering policyVerify")
	// defer fmt.Println("<-Exiting policyVerify")

	switch {
	case len(xdata.saslUsername) > 8:
		xlog.Info(fmt.Sprintf("REJECT saslUsername too long : %s ",
			xdata.saslUsername))
		return "REJECT saslUsername too long"

	case xdata.saslUsername == "" || xdata.sender == "" || xdata.clientAddress == "":
		return "REJECT missing infos"
	case blacklisted(xdata):
		xlog.Info(fmt.Sprintf("HOLD blacklisted user : %s ",
			xdata.saslUsername))
		return "HOLD blacklisted"
	case whitelisted(xdata):
		xlog.Info(fmt.Sprintf("DUNNO Whitelisted user : %s ",
			xdata.saslUsername))
		return "DUNNO"
	}

	// rcpt, _ := strconv.ParseFloat(xdata.recipientCount, 64)
	rcpt, _ := strconv.ParseInt(xdata.recipientCount, 0, 64)

	xmutex.Lock()
	defer xmutex.Unlock()

	db.Ping()
	defer db.Exec("COMMMIT")

	//  Add new entry first, ensuring correct SUM
	xlog.Info(fmt.Sprintf("Updating db: %s/%s/%s/%s",
		xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount))

	// db.Exec("set time_zone='+00:00'") // timezone UTC
	// db.Exec("set SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED ")
	// db.Exec("START TRANSACTION")

	_, err := db.Exec("INSERT INTO events SET sasl_username=?, sender=?, client_address=?, recipient_count=?",
		xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount)
	if err != nil {
		xlog.Err("ERROR while UPDATING db :" + err.Error())
		time.Sleep(3 * time.Second) // Mutex + delay = secure mysql primary key
		xlog.Info("Rate limiting similar requests, sleeping for a few secs...")
	}

	// db.Exec("COMMMIT")
	// db.Exec("set SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED ")

	// db.Exec("START TRANSACTION")

	sumerr := db.QueryRow("SELECT SUM(recipient_count) FROM events WHERE sasl_username=? AND ts>DATE_SUB(CURRENT_TIMESTAMP(3), INTERVAL 1 DAY)", xdata.saslUsername).Scan(&dbSum)

	if sumerr != nil { // Pas normal, l'erreur noRow.
		// xlog.Err("Erreur apres SUM " + sumerr.Error())
		// fmt.Printf("%v %s", sumerr, sumerr)
		// lets consider it's a new entry.
		// when message is  "converting NULL to int64 is unsupported"
		dbSum = 0
	}

	if *xdebug {
		fmt.Printf("SELECT SUM Found %s: %v recipients\n", xdata.saslUsername, dbSum)
	}

	switch {
	case dbSum+rcpt >= 2*defaultQuota:
		xlog.Info(fmt.Sprintf("REJECTING overquota (%v>2x%v) for user %s using %s from ip [%s]",
			dbSum+rcpt, defaultQuota, xdata.saslUsername, xdata.sender, xdata.clientAddress))
		return "REJECT max quota exceeded"

	case dbSum+rcpt >= defaultQuota:
		xlog.Info(fmt.Sprintf("DEFERRING overquota (%v>%v) for user %s using %s from ip [%s]",
			dbSum+rcpt, defaultQuota, xdata.saslUsername, xdata.sender, xdata.clientAddress))
		return "HOLD quota exceeded"
	default:
		return "DUNNO" // do not send OK, so we can pipe more checks in postfix
	}
}

func whitelisted(xdata connData) bool {
	if inwhitelist[xdata.saslUsername] || inwhitelist[xdata.sender] || inwhitelist[xdata.clientAddress] {
		return true
	}
	return false
}
func blacklisted(xdata connData) bool {
	if inblacklist[xdata.saslUsername] || inblacklist[xdata.sender] || inblacklist[xdata.clientAddress] {
		return true
	}
	return false
}
