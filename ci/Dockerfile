# we use ubuntu 19.04 since it ships a fairly recent postgres
# we can't use alpine for this because of a bug in -race on musl
FROM ubuntu:19.04
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
	git \
	wget \
	postgresql-client \
	build-essential \
	sqlite3

ARG GO_VERSION=1.12.7
RUN wget --quiet "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz" \
	&& echo "66d83bfb5a9ede000e33c6579a91a29e6b101829ad41fffb5c5bb6c900e109d9  go$GO_VERSION.linux-amd64.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz

ENV GOPATH=/go PATH="/go/bin:$PATH:/usr/local/go/bin"
RUN go get github.com/golangci/golangci-lint/cmd/golangci-lint
RUN go get golang.org/x/tools/cmd/cover
RUN go get github.com/mattn/goveralls
RUN go get github.com/wadey/gocovmerge
RUN go install $GOPATH/src/github.com/wadey/gocovmerge
RUN go get github.com/rubenv/sql-migrate/...

# # For once we switch to modules; ensure that golangci-lint is vendored first.
# - go install -mod vendor github.com/golangci/golangci-lint/cmd/golangci-lint
#
# # These are not necessary unless we start doing code generation during CI (which is not a terrible idea; the build
# # should fail if any generated code has not been regenerated to reflect changes).
# - go get -u github.com/rubenv/sql-migrate/...
# - go get -u github.com/volatiletech/sqlboiler/...
# - go get -u github.com/volatiletech/sqlboiler-sqlite3/...
# - go get -u github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql/...

WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash

ENV PSQL_USER=postgres PSQL_SSLMODE=disable
