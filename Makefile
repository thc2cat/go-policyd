NAME= $(notdir $(shell pwd))
TAG=$(shell git tag)

${NAME}:
	go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ${NAME}-${TAG}
