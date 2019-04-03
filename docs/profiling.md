# Protocol profiling

## Gathering timing information from the client

Requires a running test network such as the one started by `docker-compose up`.

```
go build -o bin/cachecash-curl ./cmd/cachecash-curl && ./bin/cachecash-curl cachecash://localhost:8080/file0.bin 2>&1 | grep '^\{' | jq -c 'select(.kind=="timing") | del(.time, .level, .kind)'
```

This should produce output that looks like

```
{"msg":"getBlock","when":["2019-04-03T08:50:48.85573257-07:00","2019-04-03T08:50:48.857004708-07:00"]}
{"msg":"exchangeTicketL1","when":["2019-04-03T08:50:48.857045533-07:00","2019-04-03T08:50:48.857415144-07:00"]}
{"msg":"solvePuzzle","when":["2019-04-03T08:50:48.857755336-07:00","2019-04-03T08:50:48.858876709-07:00"]}
{"msg":"exchangeTicketL2","when":["2019-04-03T08:50:48.858912964-07:00","2019-04-03T08:50:48.859282353-07:00"]}
{"msg":"decryptData","when":["2019-04-03T08:50:48.859305972-07:00","2019-04-03T08:50:48.859496154-07:00"]}
```
