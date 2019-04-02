# --------------------
# Build
# --------------------
#FROM golang:1.12.1-alpine3.9
FROM golang:1.12.1-stretch
#RUN apk update && apk add --no-cache build-base
WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash
COPY . .
RUN make PREFIX=/artifacts all

# --------------------
# Package
# --------------------
FROM alpine:3.9
COPY --from=0 /artifacts/bin/* /usr/local/bin/
