response=$(curl --write-out %{http_code} --silent --output /dev/null elasticsearch:9200)

until [ $response -ne "000" ]
do
   sleep 60
   response=$(curl --write-out %{http_code} --silent --output /dev/null elasticsearch:9200)
done

result=$(curl --write-out %{http_code} --user "elastic:$INITIAL_PASSWORD" -X POST "elasticsearch:9200/_security/user/$ELASTICSEARCH_USERNAME/_password?pretty" -H 'Content-Type: application/json' -d"
{
        \"password\": \"$ELASTICSEARCH_PASSWORD\"
}
")
echo $result
/usr/local/bin/kibana-docker
