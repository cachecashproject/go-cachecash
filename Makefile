BINNAMES2:=$(wildcard cmd/*)
BINNAMES:=$(BINNAMES2:cmd/%=%)
PREFIX?=.

.PHONY: $(BINNAMES)

all: $(BINNAMES)

$(BINNAMES):
	go build -o $(PREFIX)/bin/$@ ./cmd/$@
