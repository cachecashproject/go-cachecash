# we can't use alpine for this because of a bug in -race on musl
FROM cachecash/go-cachecash-build:latest
RUN apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y \
	postgresql-client \
	sqlite3

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
