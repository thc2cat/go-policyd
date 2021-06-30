package main

// History :
// 2019/09/10: tag 0.1 - compiling.
// 2019/09/12: tag 0.1 - deployed
// 2019/09/13: tag 0.3 - +pid,whitelist/blacklist
// 2019/09/13: tag 0.4 - +correction bug SUM (cast)
// 2019/09/16: tag 0.5 - +dbClean
// 2019/09/17: tag 0.6 - cut saslUsername@DOMAIN
// 2019/09/19: tag 0.61 - more logs for whitelist/blacklist
//                      - auto version with git tag
// 2019/09/23: tag 0.63 - log DBSUM too, suppress debug output.
// 2019/09/25: tag 0.7  - no more daemon/debug
// 2019/09/27: tag 0.72 - bug dbSum
// 0.73: show version when args are given
// 0.74: more infos for white/blacklisted
// 0.75: whitelisted only during workinghours, and not weekend
// 0.76: SQL INSERT modified to cure SQL potential injections
// 0.77: SQL DB.Exec recovery when DB.Ping() fail
//

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

	// Version is git tag version exported by Makefile
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

	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@/%s", cfg["dbuser"], cfg["dbpass"], cfg["dbname"]))
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

func policyVerify(x connData, db *sql.DB) string {

	var dbSum int64

	// Block WeekEnd or out of office hours

	switch {

	case len(x.saslUsername) > 8:
		xlog.Info(fmt.Sprintf("REJECT saslUsername too long: %s",
			x.saslUsername))
		return "REJECT saslUsername too long"

	case x.saslUsername == "" || x.sender == "" || x.clientAddress == "":
		return "REJECT missing infos"

	case blacklisted(x):
		xlog.Info(fmt.Sprintf("Holding blacklisted user: %s/%s/%s/%s",
			x.saslUsername, x.sender, x.clientAddress,
			x.recipientCount))
		return "HOLD blacklisted"

	case officehourswhitelisted(x):
		xlog.Info(fmt.Sprintf("skipping whitelisted user: %s/%s/%s/%s",
			x.saslUsername, x.sender, x.clientAddress,
			x.recipientCount))
		return "DUNNO"
	}

	xmutex.Lock() // Use mutex because dbcleaning may occur at the same time.
	defer xmutex.Unlock()

	dberr := db.Ping()
	if dberr != nil {
		xlog.Err("Skipping policyVerify db.Ping Error: " + dberr.Error())
		// Ref : https://github.com/go-sql-driver/mysql/issues/921
		db.Exec("SELECT NOW()") // Generate an error for db recovery
		return "DUNNO"          // always return DUNNO on error
	}

	defer db.Exec("COMMMIT")

	// use code in the form   => INSERT INTO TABLE users (fullname) VALUES (?)")
	// sould avoid entries like =>   '); DROP TABLE users; --
	// https://blog.sqreen.com/preventing-sql-injections-in-go-and-other-vulnerabilities/

	_, err := db.Exec("INSERT INTO "+cfg["policy_table"]+
		"(sasl_username,sender,client_address,recipient_count) VALUES (?,?,?,?)",
		x.saslUsername, x.sender, x.clientAddress, x.recipientCount)

	if err != nil {
		xlog.Err("ERROR while UPDATING db: " + err.Error())
		time.Sleep(3 * time.Second) // Mutex + delay = secure mysql primary key
		xlog.Info("Rate limited similar requests, sleeped for a 3 secs...")
	}

	sumerr := db.QueryRow("SELECT SUM(recipient_count) FROM "+cfg["policy_table"]+
		" WHERE sasl_username=? AND ts>DATE_SUB(CURRENT_TIMESTAMP(3), INTERVAL 1 DAY)",
		x.saslUsername).Scan(&dbSum)

	if sumerr != nil {
		//  ErrNoRow leads to "converting NULL to int64 is unsupported"
		// lets consider it's a new entry.
		dbSum = 0
	}

	//  Add new entry first, ensuring correct SUM
	xlog.Info(fmt.Sprintf("Updating db: %s/%s/%s/%s/%v",
		x.saslUsername, x.sender, x.clientAddress,
		x.recipientCount, dbSum))

	switch {
	case dbSum >= 2*defaultQuota:
		xlog.Info(fmt.Sprintf("REJECTING overquota (%v>2x%v) for user %s using %s from ip [%s]",
			dbSum, defaultQuota, x.saslUsername, x.sender,
			x.clientAddress))
		return "REJECT max quota exceeded"

	case dbSum >= defaultQuota:
		xlog.Info(fmt.Sprintf("DEFERRING overquota (%v>%v) for user %s using %s from ip [%s]",
			dbSum, defaultQuota, x.saslUsername, x.sender,
			x.clientAddress))
		return "HOLD quota exceeded"

	default:
		return "DUNNO" // do not send OK, so we can pipe more checks in postfix
	}
}

// Check officeours only whitelisting
func officehourswhitelisted(x connData) bool {
	var officehours, weekend bool

	if h, _, _ := time.Now().Clock(); h >= 7 && h <= 19 {
		officehours = true
	}
	if d := int(time.Now().Weekday()); d == 7 || d == 0 {
		weekend = true
	}
	return officehours && !weekend && whitelisted(x)
}

func whitelisted(d connData) bool {
	if inwhitelist[d.saslUsername] ||
		inwhitelist[d.sender] ||
		inwhitelist[d.clientAddress] {
		return true
	}
	return false
}
func blacklisted(d connData) bool {
	if inblacklist[d.saslUsername] ||
		inblacklist[d.sender] ||
		inblacklist[d.clientAddress] {
		return true
	}
	return false
}

// dbClean delete 7 days old entries in db every 24h.
func dbClean(db *sql.DB) {
	for {
		xmutex.Lock()
		err := db.Ping()
		if err == nil {
			// Keep 7 days in db
			db.Exec("DELETE from " + cfg["policy_table"] +
				" where ts<SUBDATE(CURRENT_TIMESTAMP(3), INTERVAL 7 DAY)")
		} else {
			xlog.Err("dbClean db.Exec error :" + err.Error())
		}
		xmutex.Unlock()
		// Clean every day
		time.Sleep(24 * time.Hour)
	}
}
