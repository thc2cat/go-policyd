include ../make/Makefile-for-go.mk


REMOTE_DESTINATION=root@smtps.uvsq.fr:/local/bin/

NAME= $(notdir $(shell pwd))
TAG=$(shell git tag)

build:
	 go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ${NAME}-${TAG}

dep:
	dep ensure --update

