#!/bin/sh
set -xe

case "$BUILD_MODE" in
	test)
		# run unit and sqlboiler tests
		docker run --rm --network=cachecash cachecash-ci go test -v -race -tags=sqlboiler_test ./...

		# Linting is non-fatal right now.  See `.golangci.yml` for configuration.
		docker run --rm --network=cachecash cachecash-ci golangci-lint run
		;;
	docker)
		docker-compose build
		;;
	e2e)
		ci/run-e2e.sh
		;;
esac
