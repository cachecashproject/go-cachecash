# --------------------
# Depends
# --------------------
FROM docker.elastic.co/beats/filebeat:6.6.2

# --------------------
# Build
# --------------------
#FROM golang:1.12.1-alpine3.9
FROM golang:1.12.1-stretch
#RUN apk update && apk add --no-cache build-base
WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash
COPY . .
RUN make PREFIX=/artifacts all
