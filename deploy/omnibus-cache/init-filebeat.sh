#!/bin/sh

set -euf

set +u
if test "x${ELASTICSEARCH_URL-}" = "x"; then
    # Prevent filebeat from starting if we don't know where it should forward logs.
    echo >&2 'No ELASTICSEARCH_URL set; will not forward logs.'
    touch /etc/service/filebeat/down
else
    echo >&2 "Forwarding logs to: ${ELASTICSEARCH_URL}"
    sed 's#{{ELASTICSEARCH_URL}}#'"${ELASTICSEARCH_URL}"'#g' /etc/filebeat.yml.tpl > /etc/filebeat.yml
fi
set -u
