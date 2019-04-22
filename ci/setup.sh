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
		docker network create cachecash || true
		docker run -d -p 5432:5432 -e POSTGRES_DB=publisher --name publisher-db --net=cachecash postgres:11
		docker build -t cachecash-ci -f ci/Dockerfile .

		# wait until the database is up
		while ! docker run --rm --net=cachecash cachecash-ci psql 'host=publisher-db port=5432 user=postgres dbname=publisher sslmode=disable' -c 'select 1;'; do sleep 10; done

		# apply migrations
		docker run --rm --net=cachecash cachecash-ci sql-migrate up -config=publisher/migrations/dbconfig.yml -env=docker-tests
		;;
	docker)
		update_docker_compose
		;;
	e2e)
		update_docker_compose
		docker-compose build
		;;
esac
