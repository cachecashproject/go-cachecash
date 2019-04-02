# go-cachecash

This image contains binaries for all of the programs in this repository (with entry points in `cmd/`).  They are copied
from this image to other images such as the `omnibus-cache` image in order to improve build times.

## Building

This image must be built from the root of the repository by running

```
docker build -f deploy/go-cachecash/Dockerfile -t cachecash/go-cachecash .
```
