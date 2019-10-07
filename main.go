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
// 2019/09/23 :  tag 0.63 - log DBSUM too, suppress debug output.
// 2019/09/25 : tag 0.7 - no more daemon/debug
// 2019/09/27 : tag 0.72 - bug dbSum
// 0.73 : show version when args are given
// 0.74 : more infos for white/blacklisted
//
// TODO : with context for DB blackout.

import (
	"bufio"
	"database/sql"
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
	xmutex       sync.Mutex
	defaultQuota int64

	// Version is git tag version exported by Makefile, printed by -debug
	Version string
)

const (
	syslogtag = "policyd"
	cfgfile   = "/etc/postfix/" + syslogtag + ".cfg"
)

func main() {

	if len(os.Args) > 1 {
		fmt.Printf("Usage: %s (as daemon)", syslogtag)
		os.Exit(0)
	}

	InitCfg(cfgfile)

	defaultQuota, _ = strconv.ParseInt(cfg["defaultquota"], 0, 64)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", cfg["bind"]+":"+cfg["port"])
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	initSyslog(syslogtag)

	xlog.Info(fmt.Sprintf("%s started.", Version))
	writePidfile("/var/run/" + syslogtag + ".pid")

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
		db.Exec("DELETE from " + cfg["policy_table"] + " where ts<SUBDATE(CURRENT_TIMESTAMP(3), INTERVAL 7 DAY)")
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

	switch {
	case len(xdata.saslUsername) > 8:
		xlog.Info(fmt.Sprintf("REJECT saslUsername too long : %s ",
			xdata.saslUsername))
		return "REJECT saslUsername too long"

	case xdata.saslUsername == "" || xdata.sender == "" || xdata.clientAddress == "":
		return "REJECT missing infos"
	case blacklisted(xdata):
		xlog.Info(fmt.Sprintf("HOLD blacklisted user : %s/%s/%s/%s",
			xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount))
		return "HOLD blacklisted"
	case whitelisted(xdata):
		xlog.Info(fmt.Sprintf("DUNNO Whitelisted user : %s/%s/%s/%s",
			xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount))
		return "DUNNO"
	}

	rcpt, _ := strconv.ParseInt(xdata.recipientCount, 0, 64)

	xmutex.Lock() // Use mutex because dbcleaning may occur at the same time.
	defer xmutex.Unlock()

	dberr := db.Ping()
	if dberr != nil {
		xlog.Err("Error after db.Ping " + dberr.Error())
		return "DUNNO" // always return DUNNO on error
	}

	defer db.Exec("COMMMIT")

	_, err := db.Exec("INSERT INTO "+cfg["policy_table"]+" SET sasl_username=?, sender=?, client_address=?, recipient_count=?",
		xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount)
	if err != nil {
		xlog.Err("ERROR while UPDATING db :" + err.Error())
		time.Sleep(3 * time.Second) // Mutex + delay = secure mysql primary key
		xlog.Info("Rate limiting similar requests, sleeping for a few secs...")
	}

	sumerr := db.QueryRow("SELECT SUM(recipient_count) FROM "+cfg["policy_table"]+" WHERE sasl_username=? AND ts>DATE_SUB(CURRENT_TIMESTAMP(3), INTERVAL 1 DAY)", xdata.saslUsername).Scan(&dbSum)

	if sumerr != nil { // Pas normal, l'erreur noRow.
		// lets consider it's a new entry.
		// when message is  "converting NULL to int64 is unsupported"
		dbSum = 0
	}

	//  Add new entry first, ensuring correct SUM
	xlog.Info(fmt.Sprintf("Updating db: %s/%s/%s/%s/%v",
		xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount, dbSum))

	switch {
	case dbSum >= 2*defaultQuota:
		xlog.Info(fmt.Sprintf("REJECTING overquota (%v>2x%v) for user %s using %s from ip [%s]",
			dbSum+rcpt, defaultQuota, xdata.saslUsername, xdata.sender, xdata.clientAddress))
		return "REJECT max quota exceeded"

	case dbSum >= defaultQuota:
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
