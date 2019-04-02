# --------------------
# Depends
# --------------------
FROM docker.elastic.co/beats/filebeat:6.6.2
FROM cachecash/go-cachecash

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

RUN mkdir -p /etc/service/cache/log
ADD ./deploy/omnibus-cache/cache.sh /etc/service/cache/run
ADD ./deploy/omnibus-cache/cache.log.sh /etc/service/cache/log/run
ADD ./deploy/omnibus-cache/cache.log.config /var/log/cachecash/cache/config
# # See README.md for a note about this: we require that config be bind-mounted.
# ADD ./deploy/omnibus-cache/cache.config.json /etc/cache.config.json

RUN chmod +x /etc/my_init.d/*
RUN chmod +x /etc/service/*/run
RUN chmod +x /etc/service/*/log/run

# TODO: Still need containernet packages required on Ubuntu images.

# Clean up apt when done.
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# --------------------
