# Maintainer notes

## Driving the Build System

The `Makefile` in the repository contains most menial tasks you will need to
perform to work within it. Please review there first! Things are typically
documented here, but not always.

### Doing a build

- `make all`

### Starting and Stopping the stack of services

- `make start`
- `make stop`
- `make restart`

### Cleaning up your database

- `make clean` (will also stop docker services)

### Building the Docker images

- `make build`

### Updating the Base Image

The base image contains all the dependencies you need to build our software.

To update the image of base dependencies -- from `Dockerfile.base`, whenever
this file is modified in master, this should be run:

- `make push-base-image`

### Running lint checks

- `make lint`
- `make lint-fix` (is not always advisable, review the lint output first)

### Generating code

- `make gen`
- `make gen-docs`

### Generating go modules (`go.mod` and `go.sum`) files

- `make modules`

### Perform fuzz testing

Only supported for code which supports the go-fuzz toolkit.

- `make fuzz`
