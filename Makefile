BINNAMES2:=$(wildcard cmd/*)
BINNAMES:=$(BINNAMES2:cmd/%=%)
PREFIX?=.
GOPATH?=$(shell go env GOPATH)
# use git describe after the first release
# XXX: for building from tar balls that don't have git meta data we need a fallback
GIT_VERSION:=$(or \
	$(shell git describe --long --tags 2>/dev/null), \
	$(shell printf "0.0.0.r%s.%s" "$(shell git rev-list --count HEAD)" "$(shell git rev-parse --short HEAD)") \
)

.PHONY: $(BINNAMES) dockerfiles clean

all: $(BINNAMES)

$(BINNAMES):
	go build \
		-gcflags="all=-trimpath=${GOPATH}" \
		-asmflags="all=-trimpath=${GOPATH}" \
		-ldflags="-X github.com/cachecashproject/go-cachecash.CurrentVersion=$(GIT_VERSION)" \
		-o $(PREFIX)/bin/$@ ./cmd/$@

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
	sudo rm -vrf ./data/
	docker-compose rm -f publisher-db
