FROM golang:alpine3.6 AS build
WORKDIR /go/src/app
RUN apk add --no-cache git curl
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY . /go/src/app
RUN /go/bin/dep ensure --update
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /go/bin/policyd


FROM scratch
LABEL MAINTER=thierry.caillet@uvsq.fr
LABEL VERSION="v0.36"
COPY --from=build /go/bin/policyd /
ENTRYPOINT [ "/policyd" ]
CMD [ "-v" ]
