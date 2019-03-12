# we use ubuntu 18.10 since it ships a fairly recent postgres
# we can't use alpine for this because of a bug in -race on musl
FROM ubuntu:18.10
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
	golang \
	git \
	postgresql \
	sqlite3
ENV GOPATH=/go PATH="/go/bin:$PATH"
RUN go get github.com/rubenv/sql-migrate/...

WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash
COPY . .

ENV PSQL_USER=postgres PSQL_DBNAME=publisher PSQL_SSLMODE=disable
CMD go test -v -race -tags=sqlboiler_test ./...
