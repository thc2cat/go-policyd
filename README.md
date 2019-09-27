# policyd(go) : Postfix Policyd Rate limiter 

Le but du projet est d'avoir un démon capable de limiter le nombre d'envois de mail par utilisateur authentifié.

Un petit outil à été développé sur la base d'un démon policyd existant : [polka](https://github.com/SimoneLazzaris/polka)

Il consiste en un démon utilisant les chiffres de  "policy postfix" , et emmet la réponse DUNNO (continuer)/HOLD (conserver)/REJECT (rejeter) suivant le nombre de destinataires de mails par jour.

### Wiki : [systeme:logiciel:postfix#postfix_policyd_helper](http://wiki.dsi.uvsq.fr/systeme:logiciel:postfix#postfix_policyd_helper)

### Fonctionnalités: 
  - Stockage dans une base MYSQL(mariadb) des évenements policyd
  - vérification du nombre total de destinataires sur 24h par rapport au quota max(1500).
  - conservation des mails pour les 1500 destinataires suivants ( analyse SPAM, requeue avant whitelisting ).
  - rejet a partir de 2x quota max (3000)


## Build
```
dep status  ## Pour vérifier les dépendances de code
go build    ## pour produire le binaire
```

# Installation 

 - Le binaire est installé dans __/local/bin/policyd__, 
 - le fichier de configuration dans __/etc/postfix/policyd.cfg__ ( voir exemple.cfg )
 - le service systemd dans __/local/etc/policyd.service__

## Configuration Postfix

Ajouter dans /etc/postfix/main.cf :
```
# Testing policy protocol dump (end_of_data pour avoir nbrcpt )
smtpd_end_of_data_restrictions = check_policy_service inet:127.0.0.1:9093
```

## Lancement du démon via systemd
```
Recopier unit.service dans /local/etc/policyd.service, puis

# systemctl enable /local/etc/policyd.service
# systemctl daemon-reload
# systemctl start policyd.service
# systemctl status  policyd.service

```

## SGBD mariadb 

La table stockant les évenements d'emission postfix est construite de la maniere suivante: 

```
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

```

Un outil utilisant les datas de la table events permets de sortir le top20 des usagers ( **policyd-top20.sh**)

** Note ** : 
On utilise un DATETIME(3) pour éviter les collisions de clefs lors d'enregistrements rapprochés.
Le nettoyage des enregistrements plus vieux de 7 jours est fait quotidiennement toutes les 24 heures.

## Surveillance

Le démon une fois lancé par systemd est relancé en cas de fermeture inopinée.
Cependant le demon est aussi surveillé par monit via les lignes : 
```
check program policyd with path "/usr/bin/systemctl --quiet is-active policyd"
   if status != 0 then restart
   start program = "/usr/bin/systemctl start policyd.service"
   stop  program = "/usr/bin/systemctl stop  policyd.service"
```

## Logs

Le démon utilise syslog "daemon.ERR|INFO" pour diffuser les traces.
```
[root@smtps bin]# tail /var/log/daemon.err
Sep 23 11:08:56 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 23 11:09:03 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 25 11:19:55 smtps systemd: Failed to start Policyd go daemon for Postfix.
Sep 25 11:20:20 smtps systemd: Failed to start Policyd go daemon for Postfix.

[root@smtps bin]# tail /var/log/daemon.info
Sep 26 16:23:34 smtps policyd[18771]: Updating db: nathlarm/nathalie.larmet@uvsq.fr/193.51.39.7/1/6
Sep 26 16:24:22 smtps policyd[18771]: Updating db: anabdois/anabelle.doisy@uvsq.fr/193.51.39.96/1/46
Sep 26 16:24:27 smtps policyd[18771]: Updating db: laurolch/laure.olchowik@uvsq.fr/193.51.27.247/6/48
Sep 26 16:25:27 smtps policyd[18771]: Updating db: anabdois/anabelle.doisy@uvsq.fr/193.51.39.96/1/47
Sep 26 16:25:55 smtps policyd[18771]: Updating db: solehail/solene.haillard@uvsq.fr/192.168.5.194/1/43
Sep 26 16:26:37 smtps policyd[18771]: Updating db: sandnico/sandrine.nicourd@uvsq.fr/83.204.194.47/1/38
Sep 26 16:27:03 smtps policyd[18771]: Updating db: patrgoun/patricia.gounon@uvsq.fr/193.51.24.154/3/12
Sep 26 16:27:21 smtps policyd[18771]: Updating db: anabdois/anabelle.doisy@uvsq.fr/193.51.39.96/1/48
Sep 26 16:27:53 smtps policyd[18771]: Updating db: anabdois/anabelle.doisy@uvsq.fr/193.51.39.96/2/50
Sep 26 16:28:15 smtps policyd[18771]: Updating db: migulebe/lebert@iut-velizy.uvsq.fr/193.51.24.154/1/14

```
