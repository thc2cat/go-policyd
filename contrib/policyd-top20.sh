#!/bin/sh


if [ `whoami` != "root" ]
then
 echo "must be root !"
 exit -1
fi
## Attention au "\r" en fin de ligne si le fichier est edité sous windows...
##
##  sed $'s/\r//' -i  /etc/postfix/policyd.cfg (pour corriger)

. /etc/postfix/policyd.cfg

mysql -u $dbuser -p"$dbpass" -h $dbhost $dbname <<_EOF_
SELECT  sasl_username, SUM(recipient_count) AS total  from events where  ts>SUBDATE(CURRENT_TIMESTAMP(3), INTERVAL 1 DAY) GROUP BY sasl_username ORD
ER BY total DESC LIMIT 20;
_EOF_

## Weekly stats:
## SELECT  sasl_username, SUM(recipient_count) AS total  from events GROUP BY sasl_username ORDER BY total DESC LIMIT 20;

