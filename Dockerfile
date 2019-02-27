FROM golang:1-alpine3.8
RUN apk update && apk add --no-cache build-base
WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash
COPY . .
RUN make PREFIX=/artifacts all

FROM alpine:3.8
COPY --from=0 /artifacts/bin/* /usr/local/bin/

# --------------------
# These steps are only necessary for Containernet.  Containernet expects a version of ifconfig that supports CIDR
# notation ("ifconfig up 1.2.3.4/8") but installs ifconfig to /bin which is after /sbin (where busybox ifconfig, which
# is Alpine's default, lives) in root's PATH.  /usr/local/sbin is the first element in that PATH, so symlinking things
# there fixes the problem.
RUN apk add --no-cache bash net-tools iproute2 iputils
RUN mkdir -p /usr/local/sbin
RUN for XX in ifconfig route netstat domainname hostname ypdomainname nisdomainname; do ln -s /bin/$XX /usr/local/sbin/; done
# --------------------
