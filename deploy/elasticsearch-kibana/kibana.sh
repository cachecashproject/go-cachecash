curl --user elastic:$INITIAL_PASSWORD -X POST "localhost:9200/_security/user/$ELASTICSEARCH_USERNAME/_password?pretty" -H 'Content-Type: application/json' -d"
{
        \"password\": \"$ELASTICSEARCH_PASSWORD\"
}
"
./bin/kibana
