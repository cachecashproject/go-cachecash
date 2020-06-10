#!/bin/bash
logpiped -dsn "$KVSTORE_DATABASE" -spooldir /var/spool/logpiped -esconfig /config/esconfig.json test
