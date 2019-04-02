# Maintainer notes

## Updating Go

When a new release of Go is available, the version needs to be updated in the following places:

  - `.travis.yml` (only when a new minor release is available)
  - `deploy/go-cachecash/Dockerfile` (when any new release is available)
