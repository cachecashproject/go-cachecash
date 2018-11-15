BINNAMES2:=$(wildcard cmd/*)
BINNAMES:=$(BINNAMES2:cmd/%=%)

.PHONY: $(BINNAMES)

all: $(BINNAMES)

$(BINNAMES):
	go build -o bin/$@ ./cmd/$@
