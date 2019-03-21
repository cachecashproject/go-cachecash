#!/bin/sh

# exec /usr/local/bin/cached -config /etc/cache.config.json

#exec /usr/local/bin/test-log-generator -logFile /var/log/cachecash/cache.log

mkdir -p /var/log/cachecash
/usr/local/bin/test-log-generator -logFile /var/log/cachecash/cache.log 2>/var/log/cachecash/cache.log
