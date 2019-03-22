#!/bin/sh
set -xe

case "$BUILD_MODE" in
	test)
		docker run --rm --network=cachecash -e PSQL_HOST=publisher-db cachecash-ci
		# Linting is non-fatal right now.  See `.golangci.yml` for configuration.
		golangci-lint run || true
		;;
	docker)
		docker-compose build
		;;
	e2e)
		ci/run-e2e.sh
		;;
esac
