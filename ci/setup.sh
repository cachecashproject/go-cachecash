#!/bin/bash
set -xe

update_docker_compose() {
  DOCKER_COMPOSE_VERSION=1.27.4

  sudo rm /usr/local/bin/docker-compose
  curl -L https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-`uname -s`-`uname -m` > docker-compose
  chmod +x docker-compose
  sudo mv docker-compose /usr/local/bin
}

start_db() {
  while ! docker run --rm --net=cachecash postgres:11 psql 'host=ledger-db port=5432 user=cachecash password=ledgerDB dbname=ledger sslmode=disable' -c 'select 1;'; do echo "database set up failed!"; done
  while ! docker run --rm --net=cachecash postgres:11 psql 'host=publisher-db port=5432 user=cachecash password=publisherDB dbname=publisher sslmode=disable' -c 'select 1;'; do echo "database set up failed!"; done
  while ! docker run --rm --net=cachecash postgres:11 psql 'host=kvstore-db port=5432 user=cachecash password=kvstoreDB dbname=kvstore sslmode=disable' -c 'select 1;'; do echo "database set up failed!"; done
}

# check the build image's dockerfile; if it's been altered, build it.
if git diff --stat master..HEAD | head -n -1 | awk '{ print $1 }' | grep -q Dockerfile.base 
then
  make base-image
else
  make pull-base-image
fi

make dockerfiles modules gen
if ! git diff --quiet; then
  git diff --stat
  echo 'ERROR: Generated files need to be regenerated'
  exit 1
fi

case "$BUILD_MODE" in
  test)
    docker network create cachecash --opt com.docker.network.bridge.enable_ip_masquerade=false || true
    time docker run -d -p 5433:5432 --net=cachecash --name=ledger-db --env-file ./deploy/secrets/ledger.secret postgres:11
    time docker run -d -p 5434:5432 --net=cachecash --name=publisher-db --env-file ./deploy/secrets/publisher.secret postgres:11
    time docker run -d -p 5435:5432 --net=cachecash --name=kvstore-db --env-file ./deploy/secrets/kvstore.secret postgres:11
    time docker build -t cachecash-ci ci

    # wait until the databases are up
    time start_db
    sleep 2m
    # apply migrations
    time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --net=cachecash cachecash-ci sql-migrate up -config=publisher/migrations/dbconfig.yml -env=docker-tests
    time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash --rm --net=cachecash cachecash-ci sql-migrate up -config=ledgerservice/migrations/dbconfig.yml -env=docker-tests
    ;;
  e2e)
    update_docker_compose
    time make build
    ;;
esac
