FROM golang:1.12-alpine
LABEL version="1.0"

WORKDIR $GOPATH/src/backend-reg

ADD . .


RUN apk add --no-cache git bash gcc 

ENV GO111MODULE=on

RUN go mod download
