#!/bin/sh
set -xe

# Run from the top directory

case "$BUILD_MODE" in
	test)
		# run unit and sqlboiler tests
		# each sqlboiler suite is in its own invocation to connect to a
		# separate DB and DB server because the image forces PSQL
		# endpoints today.
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --network=cachecash cachecash-ci go test -v -race ./...
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --network=cachecash cachecash-ci go test -v -race -tags=sqlboiler_test ./cache/...
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --network=cachecash cachecash-ci go test -v -race -tags=sqlboiler_test ./bootstrap/...
		time docker run -e PSQL_HOST=publisher-db -e PSQL_USER=postgres -e PSQL_DBNAME=publisher -e PSQL_SSLMODE=disable -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --network=cachecash cachecash-ci go test -v -race -tags=sqlboiler_test ./publisher/...
		time docker run -e PSQL_HOST=ledger-db -e PSQL_USER=postgres -e PSQL_DBNAME=ledger -e PSQL_SSLMODE=disable -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --network=cachecash cachecash-ci go test -v -race -tags=sqlboiler_test ./ledgerservice/...

		# Linting is non-fatal right now.  See `.golangci.yml` for configuration.
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --network=cachecash cachecash-ci golangci-lint run
		;;
	docker)
		docker-compose build
		;;
	e2e)
		ci/run-e2e.sh
		;;
esac
