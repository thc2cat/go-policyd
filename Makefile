include ../make/Makefile-for-go.mk

dtest:
	nc localhost 9093 < tests/test.data

REMOTE_DESTINATION=root@smtps.uvsq.fr:/local/bin/

