#!/bin/bash
set -xe

update_docker_compose() {
	DOCKER_COMPOSE_VERSION=1.24.0

	sudo rm /usr/local/bin/docker-compose
	curl -L https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-`uname -s`-`uname -m` > docker-compose
	chmod +x docker-compose
	sudo mv docker-compose /usr/local/bin
}

start_db() {
	while ! docker run --rm --net=cachecash postgres:11 psql 'host=ledger-db port=5432 user=postgres dbname=ledger sslmode=disable' -c 'select 1;'; do sleep 10; done
	while ! docker run --rm --net=cachecash postgres:11 psql 'host=publisher-db port=5432 user=postgres dbname=publisher sslmode=disable' -c 'select 1;'; do sleep 10; done
	while ! docker run --rm --net=cachecash postgres:11 psql 'host=kvstore-test port=5432 user=postgres dbname=kvstore sslmode=disable' -c 'select 1;'; do sleep 10; done
}

make dockerfiles
if ! git diff --quiet; then
	echo 'ERROR: Dockerfiles need to be regenerated'
	exit 1
fi

case "$BUILD_MODE" in
	test)
		docker network create cachecash --opt com.docker.network.bridge.enable_ip_masquerade=false || true
		time docker run -d -p 5433:5432 -e POSTGRES_DB=ledger --name ledger-db --net=cachecash postgres:11
		time docker run -d -p 5434:5432 -e POSTGRES_DB=publisher --name publisher-db --net=cachecash postgres:11
		time docker run -d -p 5435:5432 -e POSTGRES_DB=kvstore --name kvstore-test --net=cachecash postgres:11
		time docker build -t cachecash-ci ci

		# wait until the databases are up
		time start_db

		# apply migrations
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --net=cachecash cachecash-ci sql-migrate up -config=publisher/migrations/dbconfig.yml -env=docker-tests
		time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --net=cachecash cachecash-ci sql-migrate up -config=ledgerservice/migrations/dbconfig.yml -env=docker-tests
		;;
	docker)
		update_docker_compose
		;;
	e2e)
		update_docker_compose
		time docker-compose build
		;;
esac
