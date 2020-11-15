#!/bin/sh
set -xe

# Run from the top directory

run_test() {
  docker run --rm --network=cachecash \
    -e GOPROXY=off -e GO111MODULE=on -e PSQL_HOST="$PSQL_HOST" -e PSQL_DBNAME="$PSQL_DBNAME" \
    -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash \
    cachecash-ci \
    /bin/bash -c "
    mkfifo /tmp/fifo-$$ &&
    grep -v 'no test files' </tmp/fifo-$$ & go test -mod=vendor -v -race "$@" >/tmp/fifo-$$ &&
    RES=$? &&
    rm /tmp/fifo-$$ &&
    exit $RES"
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
      run_test -p 1 -tags "external_test sqlboiler_test" ./kv/... ./log/server/... \
      --coverprofile=kv.prof

    time make lint

    time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash \
      --rm cachecash-ci gocovmerge *.prof > coverage.out
    # Coverage exclusions.
    # Short term hack; longer term cross referencing with generated-file
    # metadata would be better, as it would be maintenance free and not suffer
    # defects such as someone putting manual code alongside a generated file
    # and having it excluded.
    # Ignore protobuf generated files
    time sed -i '/\.pb\.go/d' coverage.out
    # Ignore sqlboiler generated packages
    time sed -i '/\/models\//d' coverage.out
    # Ignore test mock helpers: they have to implement entire interfaces even
    # if only a fraction is exercised by the test needing the mock
    time sed -i '/_mock\.go/d' coverage.out
    # Disabled while we are working in the org repo: each dev branch shows as a
    # separate project branch erroneously.
    #  -e TRAVIS_BRANCH="$TRAVIS_BRANCH" \
    time docker run -v $(pwd):/go/src/github.com/cachecashproject/go-cachecash \
      --rm -e TRAVIS_JOB_ID="$TRAVIS_JOB_ID" \
      cachecash-ci goveralls -coverprofile=coverage.out \
      -service=travis-pro -repotoken "$COVERALLS_TOKEN"
    ;;
  e2e)
    ci/run-e2e.sh
    ;;
esac
