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

