# --------------------
# Depends
# --------------------
FROM docker.elastic.co/beats/filebeat:6.6.2

# --------------------
# Build
# --------------------
#FROM golang:1.12.1-alpine3.9
FROM golang:1.12.1-stretch
#RUN apk update && apk add --no-cache build-base
WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash
COPY . .
RUN make PREFIX=/artifacts all

# --------------------
# Package
# --------------------

FROM phusion/baseimage:0.11

RUN apt-get update && apt-get install -y --no-install-recommends logrotate

# Use baseimage-docker's init system.
CMD ["/sbin/my_init"]

COPY --from=0 /usr/share/filebeat/filebeat /usr/local/bin/
COPY --from=1 /artifacts/bin/* /usr/local/bin/

RUN mkdir /etc/service/filebeat
ADD ./deploy/omnibus-cache/filebeat.sh /etc/service/filebeat/run
ADD ./deploy/omnibus-cache/filebeat.yml.tpl /etc/filebeat.yml.tpl
ADD ./deploy/omnibus-cache/init-filebeat.sh /etc/my_init.d/filebeat.sh

RUN mkdir /etc/service/cache
ADD ./deploy/omnibus-cache/cache.sh /etc/service/cache/run
# # See README.md for a note about this: we require that config be bind-mounted.
# ADD ./deploy/omnibus-cache/cache.config.json /etc/cache.config.json

RUN mkdir /etc/service/logrotate
ADD ./deploy/omnibus-cache/logrotate.sh /etc/service/logrotate/run
# TMP: Disable rotation
RUN touch /etc/service/logrotate/down

RUN chmod +x /etc/my_init.d/*
RUN chmod +x /etc/service/*/run

# TODO: Still need containernet packages required on Ubuntu images.

# Clean up apt when done.
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# --------------------
