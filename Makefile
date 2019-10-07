include ../make/Makefile-for-go.mk

REMOTE_DESTINATION=root@smtps.uvsq.fr:/local/bin/

NAME= $(notdir $(shell pwd))
TAG=$(shell git tag)

<<<<<<< HEAD
build:
	@go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ${NAME}-${TAG}
	@notify-send 'Build Complete' 'Your project has been build successfully!' -u normal -t 7500 -i checkbox-checked-symbolic

release:
	scp ${NAME}-${TAG} ${REMOTE_DESTINATION}
=======
${NAME}-${TAG}:
	 go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ${NAME}-${TAG}

dep:
	dep ensure --update


deploy:	${NAME}-${TAG}
	scp ${NAME}-${TAG}  ${REMOTE_DESTINATION}
>>>>>>> Makefile deploy
