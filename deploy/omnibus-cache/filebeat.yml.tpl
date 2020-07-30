# filebeat.prospectors:
# - paths:
#    - test.log
#   input_type: log
#   json.keys_under_root: true
#   json.add_error_key: true

filebeat.inputs:
- type: log
  paths:
    - /var/log/cachecash/cache/current
    - /var/log/cachecash/cache/*.s
  json.keys_under_root: true
  json.add_error_key: true

output.elasticsearch:
  hosts: ["{{ELASTICSEARCH_URL}}"]

# output.console:
#   pretty: true
