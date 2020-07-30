##
# Questions? Read MAINTAINERS.md!
##

FUZZTAGS?=fuzz
PREFIX?=$(shell realpath .)
GOPATH?=$(shell go env GOPATH)
# use git describe after the first release
# XXX: for building from tar balls that don't have git meta data we need a fallback
GIT_VERSION:=$(or \
	$(shell git describe --long --tags 2>/dev/null), \
	$(shell printf "0.0.0.r%s.%s" "$(shell git rev-list --count HEAD)" "$(shell git rev-parse --short HEAD)") \
)

BASE_IMAGE=cachecash/go-cachecash-build:latest

# These contain the dirs that can be generated without SQL or other servers
# running. Full generation requires a docker-compose-up + then running `go
# generate ./...`, but that can then lead to tool skew issues. So the canonical
# thing to do is to do that, and then run `make gen` one last time.
GEN_DIRS=./ccmsg/... ./ledger ./log/... ./metrics/... ./blockexplorer/... ./bootstrap/... ./ranger/...
GEN_CONTAINER_DIR=/go/src/github.com/cachecashproject/go-cachecash
GEN_DOCS_FLAGS=-Iccmsg -Ilog -Imetrics
GEN_PROTO_FILES={ccmsg,log,metrics}/*.proto
GEN_DOCKER=docker run --rm -it -e GO111MODULE=on -e GOCACHE=/tmp/go-cache -w ${GEN_CONTAINER_DIR} -u $$(id -u):$$(id -g) -v ${PWD}:${GEN_CONTAINER_DIR} ${BASE_IMAGE}

.PHONY: dockerfiles clean lint lint-fix dev-setup \
	fuzz fuzz-ledger fuzz-ranger \
	gen gen-docs modules \
	base-image pull-base-image push-base-image \
	restart stop build start

all:
	GO111MODULE=on GOBIN=$(PREFIX)/bin go install -mod=vendor \
		-gcflags="all=-trimpath=${GOPATH}" \
		-asmflags="all=-trimpath=${GOPATH}" \
		-ldflags="-X github.com/cachecashproject/go-cachecash.CurrentVersion=$(GIT_VERSION)" \
		./cmd/...
ccc:
	GO111MODULE=on GOBIN=$(PREFIX)/bin go install -mod=vendor \
                -gcflags="all=-trimpath=${GOPATH}" \
                -asmflags="all=-trimpath=${GOPATH}" \
                -ldflags="-X github.com/cachecashproject/go-cachecash.CurrentVersion=$(GIT_VERSION)" \
                ./cmd/cachecash-curl/...

restart: stop build start

stop:
	docker-compose rm -f

build:
	docker-compose build

start: build
	docker-compose up

dockerfiles:
	cat deploy/dockerfiles/autogen-warning.txt \
		deploy/dockerfiles/build.stage \
		deploy/dockerfiles/filebeat.stage \
		deploy/dockerfiles/omnibus.stage > \
		Dockerfile
	cat deploy/dockerfiles/autogen-warning.txt \
		deploy/dockerfiles/build.stage > \
		deploy/dockerfiles/Dockerfile.build

clean:
	docker-compose down
	docker-compose rm -sf publisher-db ledger-db
	sudo rm -vrf ./data/

lint:
	docker build -t cachecash-ci ci
	docker run -v ${PWD}:/go/src/github.com/cachecashproject/go-cachecash --rm cachecash-ci golangci-lint run -v

lint-fix:
	docker build -t cachecash-ci ci
	docker run -v ${PWD}:/go/src/github.com/cachecashproject/go-cachecash --rm cachecash-ci golangci-lint run --fix

dev-setup:
	go get -u github.com/rubenv/sql-migrate/...
	go get -u github.com/volatiletech/sqlboiler/...
	go get -u github.com/volatiletech/sqlboiler-sqlite3/...
	go get -u github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql/...

base-image:
	docker build --pull --no-cache -t cachecash/go-cachecash-build:latest -f Dockerfile.base .

push-base-image: base-image
	docker push cachecash/go-cachecash-build:latest

pull-base-image:
	docker pull cachecash/go-cachecash-build:latest

gen: all
	@echo >&2
	@echo >&2 'If this errors, `make pull-base-image`'
	@echo >&2
	$(GEN_DOCKER) \
		go generate -mod=vendor ${GEN_DIRS}

gen-docs: pull-base-image
	mkdir -p docs-gen
	$(GEN_DOCKER) \
		bash -c "protoc --doc_out=${GEN_CONTAINER_DIR}/docs-gen --doc_opt=html,index.html ${GEN_DOCS_FLAGS} -I. -I/go/src $$(eval echo ${GEN_PROTO_FILES})"

modules:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

# Fuzz everything (with -j to get parallelism)
fuzz: fuzz-ledger fuzz-ranger

.PHONY: ledger.unfuzzed ranger.unfuzzed ranger-sample
%.fuzzed : %.unfuzzed
	mkdir -p fuzz/$(shell basename ${FUZZ})-fuzz-workdir/corpus
	go-fuzz-build -tags ${FUZZTAGS} github.com/cachecashproject/go-cachecash/${FUZZ}
	go-fuzz ${FUZZPARAMS} -bin=./$(shell basename ${FUZZ})-fuzz.zip -workdir=fuzz/$(shell basename ${FUZZ})-fuzz-workdir


fuzz-ledger: FUZZ=ledger
fuzz-ledger: ledger.fuzzed

fuzz-ranger : FUZZ=ranger/testdata/pkg
fuzz-ranger : ranger.fuzzed
ranger.fuzzed +: ranger-sample

ranger-sample:
	go test -race -v ./ranger
