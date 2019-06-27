#!/bin/sh
set -xe

# Run from the top directory

case "$BUILD_MODE" in
	test)
		# run unit and sqlboiler tests
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --network=cachecash cachecash-ci go test -v -race -tags=sqlboiler_test ./...

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
