#!/bin/sh
set -xe

# Run from the top directory

run_test() {
	docker run --rm --network=cachecash \
		-e PSQL_HOST="$PSQL_HOST" -e PSQL_DBNAME="$PSQL_DBNAME" \
		-v $(pwd):/go/src/github.com/cachecashproject/go-cachecash \
		cachecash-ci \
		go test -v -race "$@"
}

case "$BUILD_MODE" in
	test)
		# run unit and sqlboiler tests
		# each sqlboiler suite is in its own invocation to connect to a
		# separate DB and DB server because the image forces PSQL
		# endpoints today.
		rm -f *.prof
		run_test ./... --coverprofile=default.prof
		run_test -tags=sqlboiler_test ./cache/... \
			--coverprofile=cache.prof
		run_test -tags=sqlboiler_test ./bootstrap/... \
			--coverprofile=bootstrap.prof
		PSQL_HOST=publisher-db PSQL_DBNAME=publisher \
			run_test -tags=sqlboiler_test ./publisher/... \
			--coverprofile=publisher.prof
		PSQL_HOST=ledger-db PSQL_DBNAME=ledger \
			run_test -tags=sqlboiler_test ./ledgerservice/... \
			--coverprofile=ledger.prof
		PSQL_HOST=kvstore-test PSQL_DBNAME=kvstore \
			run_test -tags "external_test sqlboiler_test" ./kv/... \
			--coverprofile=kv.prof

		# Linting is non-fatal right now.  See `.golangci.yml` for configuration.
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash \
			--rm --network=cachecash cachecash-ci golangci-lint run --deadline 5m
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash \
			--rm cachecash-ci gocovmerge *.prof > coverage.out
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash \
			--rm cachecash-ci goveralls -coverprofile=coverage.out \
			-service=travis-pro -repotoken "$COVERALLS_TOKEN"
		;;
	docker)
		docker-compose build
		;;
	e2e)
		ci/run-e2e.sh
		;;
esac
