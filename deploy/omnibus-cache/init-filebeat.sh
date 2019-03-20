#!/bin/sh

set -euf -o pipefail

if [ -z "${ELASTICSEARCH_URL}" ]; then
    echo &>2 "Forwarding logs to: ${ELASTICSEARCH_URL}"
    sed 's#{{ELASTICSEARCH_URL}}#'"${ELASTICSEARCH_URL}"'#g' /etc/filebeat.yml.tpl > /etc/filebeat.yml
else
    # Prevent filebeat from starting if we don't know where it should forward logs.
    echo &>2 'No ELASTICSEARCH_URL set; will not forward logs.'
    touch /etc/service/filebeat/down
fi
