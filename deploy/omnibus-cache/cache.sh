#!/bin/sh

# exec /usr/local/bin/cached -config /etc/cache.config.json

mkdir -p /var/log/cachecash
exec /usr/local/bin/test-log-generator -logFile='-'
#/usr/local/bin/test-log-generator -logFile /var/log/cachecash/cache.log 2>/var/log/cachecash/cache.log
