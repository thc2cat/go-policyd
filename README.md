
<<<<<<< HEAD
[![Build Status](https://travis-ci.com/thc2cat/go-policyd.svg?branch=for_github)](https://travis-ci.org/thc2cat/go-policyd)
=======
[![Build Status](https://img.shields.io/travis/thc2cat/go-policyd.svg?style=flat-square)](https://travis-ci.org/thc2cat/go-policyd)
>>>>>>> + travis.yml

# ![lock](contrib/24-security-lock.png) go-policyd : Postfix Policyd Rate limiter  

<<<<<<< HEAD
go-policyd project purpose is to limit spam volume emission sent via authenticated user with postfix when phishing succeeds.
=======
go-policyd project purpose is to limit spam emission sent via authenticated user with postfix when phishing succeeds.
>>>>>>> for_github

This daemon has been written from a existing policyd daemon : [polka](https://github.com/SimoneLazzaris/polka)


go-policyd use postfix policy protocol, and based on recipients numbers cumulated by day respond DUNNO (neutral)/ HOLD (store in quarantine)/ REJECT ( refuse mail.).


### Main features: 
<<<<<<< HEAD
  ![](contrib/accept.png) Single binary serving as network daemon, allowing multiple remote docker usage.

  ![](contrib/accept.png) Mysql(Mariadb) for decentralized storage of policyd events.

  ![](contrib/accept.png) Quota of total recipients by day for an authenticated sender (max 1500 recipients).

  ![](contrib/accept.png) Hold queue when over quota  for mail analysis and requeue if whitelisting or errors.

  ![](contrib/accept.png) rejection when recipients sum is over 2x quota max (3000)

  ![](contrib/accept.png) Whitelisting is available during offices hours ( not Week Ends ). Blacklisted entries are permanent.

## Build

"go build" should download dependencies and build binary

Via Makefile : 

```Makefile
NAME= $(notdir $(shell pwd))
TAG=$(shell git tag)

{NAME}:
  go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ${NAME}
=======
  ![](contrib/accept.png) Mysql(Mariadb) storage of policyd events.

  ![](contrib/accept.png) Quota of total recipients by day for an authenticated sender (max 1500 recipients).

  ![](contrib/accept.png) Hold queue when over quota  for mail analysis and requeue if whitelisting or errors.

  ![](contrib/accept.png) rejection when recipients sum is over 2x quota max (3000)

## Build

"go build" will download dependencies and build binary

Via Makefile : 

```Makefile
NAME= $(notdir $(shell pwd))
TAG=$(shell git tag)

{NAME}:
  go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ${NAME}
```

# Setup  

 - binary  __/local/bin/policyd__, 
 - config file  __/etc/postfix/policyd.cfg__ ( see contrib/policyd.cfg )
 - CentOS systemd service  __/local/etc/policyd.service__
 - mariadb

## Postfix Configuration 

Add /etc/postfix/main.cf :
```
# Testing policy protocol dump (end_of_data pour avoir nbrcpt )
smtpd_end_of_data_restrictions = check_policy_service inet:127.0.0.1:9093
>>>>>>> for_github
```

# How to use go-policyd

<<<<<<< HEAD
 - Configure postfix for policyd restrictions
 - Create mariadb database
 - copy binary  in __/local/bin/policyd__, 
 - Adapt config file contrib/policyd.cfg to  __/etc/postfix/policyd.cfg__ 
 - Enable (CentOS systemd) service  __/local/etc/policyd.service__
=======
Whitelisting is available during offices hours ( not Week Ends)
Blacklisted entries are permanent.

## systemd daemon setup 
```Shell Session
>>>>>>> for_github

## Postfix Configuration 

Add /etc/postfix/main.cf :
```
# Policyd restrictions ( at end_of_data stage for nbrcpt )
smtpd_end_of_data_restrictions = check_policy_service inet:127.0.0.1:9093
```
Then verify configuration with 'postfix check' command

<<<<<<< HEAD
## SGBD mariadb database creation
=======
## SGBD mariadb table creation
>>>>>>> for_github

```SQL
> CREATE USER 'policyd_daemon'@'localhost' IDENTIFIED BY 'yourChoiceOfPassword';
Query OK, 0 rows affected (0.01 sec)

> CREATE DATABASE policyd;
Query OK, 1 row affected (0.00 sec)

> CREATE TABLE IF NOT EXISTS `events` (
  `ts` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `sasl_username` char(8) NOT NULL DEFAULT '',
  `sender` char(80) NOT NULL DEFAULT '',
  `client_address`char(80) NOT NULL DEFAULT '',
  `recipient_count` int(6) DEFAULT NULL,
  PRIMARY KEY (`ts`,`sasl_username`,`sender`,`client_address`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

> GRANT ALL PRIVILEGES ON policyd.* TO 'policyd_daemon'@'localhost';
Query OK, 0 rows affected (0.00 sec)
<<<<<<< HEAD
```



__Nota__ : 
  - DATETIME(3) avoid key collision when multiples connections occurs.
  - The cleaning of records older than 7 days is done daily every 24 hours.
  - A policyd top 20 usage display utility is available in **contrib/policyd-top20.sh**

## CentOS Systemd daemon setup 

```Shell Session
# mkdir -p /local/bin/
# cp policyd /local/bin
# cp contrib/unit.service /local/etc/policyd.service
# cp contrib/policyd.cfg /etc/postfix/policyd.cfg
# edit /etc/postfix/policyd.cfg

# systemctl enable /local/etc/policyd.service
# systemctl daemon-reload
# systemctl start policyd.service
# systemctl status  policyd.service
```

## Monit check (optional)

The daemon once started by systemd is restarted in the event of an unexpected shutdown.

=======
```

A policyd top20 usage display utility is available in **contrib/policyd-top20.sh**

__Nota__ : 
DATETIME(3) avoid key collision when multiples connections occurs.
The cleaning of records older than 7 days is done daily every 24 hours.

## monit check

The daemon once started by systemd is restarted in the event of an unexpected shutdown.

>>>>>>> for_github
However the daemon is also monitored by monit via the lines

```
check program policyd with path "/usr/bin/systemctl --quiet is-active policyd"
   if status != 0 then restart
   start program = "/usr/bin/systemctl start policyd.service"
   stop  program = "/usr/bin/systemctl stop  policyd.service"
```

## Logs examples

<<<<<<< HEAD
go-policyd use syslog "daemon.ERR|INFO" facility 
=======
go-policyd use syslog "daemon.ERR|INFO"  facility 
>>>>>>> for_github

```Shell Session
# tail /var/log/daemon.err
Sep 23 11:08:56 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 23 11:09:03 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 25 11:19:55 smtps systemd: Failed to start Policyd go daemon for Postfix.
Sep 25 11:20:20 smtps systemd: Failed to start Policyd go daemon for Postfix.

# tail /var/log/daemon.info
Sep 26 16:23:34 smtps policyd[18771]: Updating db: nathxxx/nathalie.xxxxxxxx@mydomain.fr/192.168.39.7/1/6
Sep 26 16:24:22 smtps policyd[18771]: Updating db: anaxxx/anabelle.xxxxxxx@mydomain.fr/192.168.39.96/1/46
Sep 26 16:28:15 smtps policyd[18771]: Updating db: migxxx/migxxxxx@sub.mydomain.fr/192.168.24.154/1/14
<<<<<<< HEAD
```
The format is identifier / email / ip / recipients / recipientssumforthelast24h.

```
Sep 26 16:27:53 smtps policyd[18771]: Updating db: anaxxx/anabelle.xxxxxxx@mydomain.fr/192.168.39.96/2/50
```
indicates the sending of an email to 2 recipients, bringing the total of recipients over 24 hours to 50.

// cSpell:ignore policyd,smtps,sasl,monit,mariadb,smtpd,inet,SGBD
// cSpell:ignore systeme,nbrcpt,Inno,nathxxx,anaxxx,migxxx
=======

```
The format is identifier / email / ip / recipients / recipientssumforthelast24h.


Sep 26 16:27:53 smtps policyd[18771]: Updating db: anaxxx/anabelle.xxxxxxx@mydomain.fr/192.168.39.96/2/50

indicates the sending of an email to 2 recipients, bringing the total of recipients over 24 hours to 50.

// cSpell:ignore policyd,smtps,sasl,monit,mariadb,smtpd,inet,SGBD
// cSpell:ignore systeme,nbrcpt,Inno,nathxxx,anaxxx,migxxx
<<<<<<< HEAD
>>>>>>> for_github
=======
>>>>>>> + travis.yml
