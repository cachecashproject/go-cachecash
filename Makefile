BINNAMES2:=$(wildcard cmd/*)
BINNAMES:=$(BINNAMES2:cmd/%=%)
PREFIX?=.
GOPATH?=$(shell go env GOPATH)

.PHONY: $(BINNAMES)

all: $(BINNAMES)

$(BINNAMES):
	go build \
		-gcflags="all=-trimpath=${GOPATH}" \
		-asmflags "all=-trimpath=${GOPATH}" \
		-o $(PREFIX)/bin/$@ ./cmd/$@
