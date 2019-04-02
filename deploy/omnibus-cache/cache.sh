#!/bin/sh

# # If testing the log rotation and collection mechanism, this may be useful.
# exec /usr/local/bin/test-log-generator -logFile='-'

# Output logs to stdout where svlogd can collect them.
exec /usr/local/bin/cached -config /etc/cache.config.json  -logFile='-'

