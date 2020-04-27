FROM golang:1.14.2-alpine3.11

COPY ./ /go/src/github.com/bluedresscapital/coattails

WORKDIR /go/src/github.com/bluedresscapital/coattails

RUN go build ./cmd/coattails/main.go

