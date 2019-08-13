PREFIX?=$(shell realpath .)
GOPATH?=$(shell go env GOPATH)
# use git describe after the first release
# XXX: for building from tar balls that don't have git meta data we need a fallback
GIT_VERSION:=$(or \
	$(shell git describe --long --tags 2>/dev/null), \
	$(shell printf "0.0.0.r%s.%s" "$(shell git rev-list --count HEAD)" "$(shell git rev-parse --short HEAD)") \
)

GEN_PROTO_DIRS=./ccmsg/... ./log/... ./metrics/...
GEN_CONTAINER_DIR=/go/src/github.com/cachecashproject/go-cachecash
GEN_PROTO_FILE=${GEN_CONTAINER_DIR}/ccmsg/cachecash.proto 
GEN_DOCKER=docker run -it -w ${GEN_CONTAINER_DIR} -u $$(id -u):$$(id -g) -v ${PWD}:${GEN_CONTAINER_DIR} cachecash-gen

.PHONY: dockerfiles clean lint lint-fix \
	dev-setup gen gen-image gen-docs modules

all:
	GO111MODULE=on GOBIN=$(PREFIX)/bin go install -mod=vendor \
		-gcflags="all=-trimpath=${GOPATH}" \
		-asmflags="all=-trimpath=${GOPATH}" \
		-ldflags="-X github.com/cachecashproject/go-cachecash.CurrentVersion=$(GIT_VERSION)" \
		./cmd/...

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
	docker run -v ${PWD}:/go/src/github.com/cachecashproject/go-cachecash --rm cachecash-ci golangci-lint run

lint-fix:
	docker build -t cachecash-ci ci
	docker run -v ${PWD}:/go/src/github.com/cachecashproject/go-cachecash --rm cachecash-ci golangci-lint run --fix

dev-setup:
	go get -u github.com/rubenv/sql-migrate/...
	go get -u github.com/volatiletech/sqlboiler/...
	go get -u github.com/volatiletech/sqlboiler-sqlite3/...
	go get -u github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql/...

gen: gen-image
	$(GEN_DOCKER) \
		go generate ${GEN_PROTO_DIRS}

gen-docs:
	mkdir -p docs-gen
	$(GEN_DOCKER) \
		protoc --doc_out=${GEN_CONTAINER_DIR}/docs-gen --doc_opt=html,index.html --proto_path=/go/src ${GEN_PROTO_FILE}

gen-image:
	docker build -t cachecash-gen -f Dockerfile.gen .

modules:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor
