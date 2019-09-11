package main

// History :
// 2019/09/10 :  tag 0.1 - compiling.
//

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type connData struct {
	saslUsername   string
	clientAddress  string
	sender         string
	recipientCount string
}

var (
	xdebug *bool
	xmutex sync.Mutex
)

func init() {
	xdebug = flag.Bool("debug", false, "enable debugging")
}

func main() {
	InitCfg()
	flag.Parse()
	if !*xdebug {
		daemon(0, 0)
	} else {
		fmt.Println("Starting in debug mode")
	}
	// Listen for incoming connections.
	l, err := net.Listen("tcp", cfg["bind"]+":"+cfg["port"])
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	// open connection to the database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?autocommit=false", cfg["dbuser"], cfg["dbpass"], cfg["dbhost"], cfg["dbport"], cfg["dbname"]))
	if err != nil {
		fmt.Println("ERROR CONNECTING MYSQL")
		os.Exit(1)
	}
	defer db.Close()

	initSyslog()

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			xlog.Err("Error accepting: " + err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn, db)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, db *sql.DB) {
	var xdata connData
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
			fmt.Println("Error processing line")
			continue
		}
		// if (*xdebug) { fmt.Println("..", s, ":", vv[0],"->",vv[1]) }
		vv[0] = strings.Trim(vv[0], " \n\r")
		vv[1] = strings.Trim(vv[1], " \n\r")
		switch vv[0] {
		case "sasl_username":
			xdata.saslUsername = vv[1]
		case "sender":
			xdata.sender = vv[1]
		case "client_address":
			xdata.clientAddress = vv[1]
		case "recipient_count":
			xdata.recipientCount = vv[1]
		}
	}
	db.Ping()
	db.Exec("set time_zone='+00:00'") // timezone UTC
	db.Exec("set session TRANSACTION ISOLATION LEVEL REPEATABLE READ")
	resp := policyVerify(xdata, db)
	conn.Write([]byte(fmt.Sprintf("action=%s\n\n", resp)))
	conn.Close()
}

func policyVerify(xdata connData, db *sql.DB) string {

	var defaultQuota, dbSum float64

	if blacklisted(xdata) {
		return "HOLD blacklisted"
	}
	if whitelisted(xdata) {
		return "DUNNO"
	}

	if xdata.saslUsername == "" || xdata.sender == "" || xdata.clientAddress == "" {
		return "REJECT missing infos"
	}

	defaultQuota, _ = strconv.ParseFloat(cfg["defaultquota"], 64)
	rcpt, _ := strconv.ParseFloat(xdata.recipientCount, 64)

	xmutex.Lock()
	defer xmutex.Unlock()

	defer db.Exec("COMMIT") //defer db.Exec("COMMIT; UNLOCK TABLES;")
	db.Exec("START TRANSACTION")

	// err := db.QueryRow("SELECT max, quota, unix_timestamp(ts), unix_timestamp(now()) FROM "+cfg["policy_table"]+" where type=? and item=? FOR UPDATE", xtype, xitem).Scan(&mx, &quota, &ts, &sNow)
	err := db.QueryRow("SELECT SUM(rcpt_count) FROM events where sasl_username=? and ts>TIME(DATE_SUB(NOW(), INTERVAL 1 DAY)) FOR UPDATE", xdata.saslUsername).Scan(&dbSum)
	switch {

	// Creation d'une entree  :
	case err == sql.ErrNoRows:
		if *xdebug {
			fmt.Println("NOT FOUND")
		}
		xlog.Info("New entry for username " + xdata.saslUsername + " +" + xdata.recipientCount)

		_, err = db.Exec("INSERT INTO events set ts=now(), sasl_username=?, sender=?, client_address=?, recipient_count=?",
			xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount)
		if err != nil {
			xlog.Err(err.Error())
			if *xdebug {
				fmt.Println("ERROR INSERTING:", err.Error())
			}
		}

		// J'ai une erreur ???
	case err != nil:
		xlog.Err("ERROR: " + err.Error())
		return "DUNNO"
	default:
		if *xdebug {
			fmt.Printf("FOUND for (%s) : %f recipients\n", xdata.saslUsername, dbSum)
		}

	}

	if dbSum+rcpt > defaultQuota {
		xlog.Info(fmt.Sprintf("DEFERRING overquota (%v>%v) for user %s using %s from ip [%s]", dbSum, defaultQuota, xdata.saslUsername, xdata.sender, xdata.clientAddress))
		return "HOLD quota exceeded"
	}

	// Still there ? ok. Add new entry
	_, err = db.Exec("INSERT INTO events set ts=now(), sasl_username=?, sender=?, client_address=?, recipient_count=?",
		xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount)
	if err != nil {
		xlog.Err(err.Error())
		if *xdebug {
			fmt.Println("ERROR UPDATING:", err.Error())
		}
	}
	xlog.Info(fmt.Sprintf("Updating events for user (%s) using <%s> from ip [%s] for %s recipients",
		xdata.saslUsername, xdata.sender, xdata.clientAddress, xdata.recipientCount))

	return "DUNNO" // not OK so we can pipe more checks in postfix

}

func whitelisted(xdata connData) bool {
	return false
}
func blacklisted(xdata connData) bool {
	return false
}
