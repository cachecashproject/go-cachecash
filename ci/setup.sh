#!/bin/sh
set -xe

update_docker_compose() {
	DOCKER_COMPOSE_VERSION=1.24.0

	sudo rm /usr/local/bin/docker-compose
	curl -L https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-`uname -s`-`uname -m` > docker-compose
	chmod +x docker-compose
	sudo mv docker-compose /usr/local/bin
}

case "$BUILD_MODE" in
	test)
		go get github.com/golangci/golangci-lint/cmd/golangci-lint
		go get github.com/rubenv/sql-migrate/...

		docker network create cachecash
		docker run -d -p 5432:5432 -e POSTGRES_DB=publisher --name publisher-db --net=cachecash postgres:10
		docker build -t cachecash-ci -f ci/Dockerfile .

		while ! psql 'host=127.0.0.1 port=5432 user=postgres dbname=publisher sslmode=disable' -c 'select 1;'; do sleep 10; done
		(cd publisher/migrations/; sql-migrate up -config=dbconfig.yml -env=development)

		# # For once we switch to modules; ensure that golangci-lint is vendored first.
		# - go install -mod vendor github.com/golangci/golangci-lint/cmd/golangci-lint
		#
		# # These are not necessary unless we start doing code generation during CI (which is not a terrible idea; the build
		# # should fail if any generated code has not been regenerated to reflect changes).
		# - go get -u github.com/rubenv/sql-migrate/...
		# - go get -u github.com/volatiletech/sqlboiler/...
		# - go get -u github.com/volatiletech/sqlboiler-sqlite3/...
		# - go get -u github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql/...
		;;
	docker)
		update_docker_compose
		;;
	e2e)
		update_docker_compose
		docker-compose build
		;;
esac
