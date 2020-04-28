FROM golang:1.14.2-alpine3.11

RUN apk update update && apk add \
  curl \
  postgresql \
  postgresql-contrib

COPY ./ /go/src/github.com/bluedresscapital/coattails

WORKDIR /go/src/github.com/bluedresscapital/coattails

RUN go build ./cmd/coattails/main.go

