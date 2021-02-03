
[![pipeline status](https://git.dsi.uvsq.fr/thiecail/policyd/badges/deployed/pipeline.svg)](https://git.dsi.uvsq.fr/thiecail/policyd/commits/deployed)

# ![lock](contrib/24-security-lock.png) policyd(go) : Postfix Policyd Rate limiter  

Ce projet a pour but de limiter le réduire le volume de spam envoyés via postfix par utilisateur authentifié lors de phishing réussis.

Un petit outil à été développé sur la base d'un démon policyd existant : [polka](https://github.com/SimoneLazzaris/polka)

Il consiste en un démon utilisant les chiffres de  "policy postfix" , et émet la réponse DUNNO (continuer)/HOLD (conserver)/REJECT (rejeter) suivant le nombre de destinataires de mails par jour.

### Wiki :    [systeme:logiciel:postfix#postfix_policyd_helper](http://wiki.dsi.uvsq.fr/systeme:logiciel:postfix#postfix_policyd_helper) 

### Fonctionnalités : 
  ![](contrib/accept.png) Stockage dans une base MYSQL(mariadb) des évènements policyd

  ![](contrib/accept.png) vérification du nombre total de destinataires sur 24h par rapport au quota max(1500).

  ![](contrib/accept.png) conservation des mails pour les 1500 destinataires suivants ( pour analyse , et requeue si  whitelisting ).

  ![](contrib/accept.png) rejet a partir de 2x quota max (3000)

## Build
```
go build         ## pour créer le binaire
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

## Whitelisting

Le whitelisting est permis durant les heures de travail ( 7h-19h , en dehors des WE ).
Une entrée blacklistée est permanente.

## Lancement du démon via systemd
```Shell Session
Recopier unit.service dans /local/etc/policyd.service, puis

# cp contrib/unit.service /local/etc/policyd.service
# systemctl enable /local/etc/policyd.service
# systemctl daemon-reload
# systemctl start policyd.service
# systemctl status  policyd.service

```

## SGBD mariadb 

La table stockant les évènements d'emission postfix est construite de la manière suivante: 

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

```

Un outil utilisant les datas de la table events permets de sortir le top20 des usagers ( **contrib/policyd-top20.sh**)

__Note__ : 
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

```Shell Session
# tail /var/log/daemon.err
Sep 23 11:08:56 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 23 11:09:03 smtps policyd[8168]: ERROR while UPDATING db :dial tcp 127.0.0.1:3306: connect: connection refused
Sep 25 11:19:55 smtps systemd: Failed to start Policyd go daemon for Postfix.
Sep 25 11:20:20 smtps systemd: Failed to start Policyd go daemon for Postfix.

# tail /var/log/daemon.info
Sep 26 16:23:34 smtps policyd[18771]: Updating db: nathxxxx/nathalie.xxxxxxxx@mydomain.fr/192.168.39.7/1/6
Sep 26 16:24:22 smtps policyd[18771]: Updating db: anaxxxxxx/anabelle.xxxxxxx@mydomain.fr/192.168.39.96/1/46
Sep 26 16:28:15 smtps policyd[18771]: Updating db: migulebe/migxxxxx@sub.mydomain.fr/192.168.24.154/1/14

```

Le format est identifiant/email/ip/destinataires/sumdestinataire24h.

Par exemple: la ligne 

Sep 26 16:27:53 smtps policyd[18771]: Updating db: anaxxxxxx/anabelle.xxxxxxx@mydomain.fr/192.168.39.96/2/50

indique l'envoi d'un mail vers 2 destinataires, portant le total de destinataires sur 24h à 50.

// cSpell:ignore policyd,smtps,sasl,monit,mariadb,smtpd,inet,SGBD