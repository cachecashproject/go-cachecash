FROM golang:1-alpine3.8
RUN apk update && apk add --no-cache build-base
WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash
COPY . .
RUN make PREFIX=/artifacts all

FROM alpine:3.8
COPY --from=0 /artifacts/bin/* /usr/local/bin/
