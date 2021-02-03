FROM golang:alpine AS build
WORKDIR /go/src/app
COPY . /go/src/app
RUN go mod download
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /go/bin/policyd

FROM scratch
LABEL MAINTENER=thierry.caillet@uvsq.fr
LABEL VERSION="v0.36"
COPY --from=build /go/bin/policyd /
ENTRYPOINT [ "/policyd" ]
CMD [ "-v" ]
