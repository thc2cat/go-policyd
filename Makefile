include ../make/Makefile-for-go.mk

dtest:
	nc localhost 9093 < tests/test.data

REMOTE_DESTINATION=root@smtps.uvsq.fr:/local/bin/

GOLDFLAGS += -X main.Version=${NAME}-${TAG}
GOFLAGS = -ldflags "$(GOLDFLAGS)"

release:	${NAME}-${TAG}
	goupx -9 ${NAME}-${TAG}

do-release:	${NAME}-${TAG}
	scp ${NAME}-${TAG} ${REMOTE_DESTINATION}

changelog:
	gitchangelog > Changelog
