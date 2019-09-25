# Maintainer notes

## Updating the Base Image

The base image contains all the dependencies you need to build our software.

To update the image of base dependencies -- from `Dockerfile.base`, whenever
this file is modified in master, this should be run:

- `make push-base-image`
