# Protocol profiling

## Via OpenCensus / Jaeger

Requires a running test network with Jaeger tracing enabled and an accessible
Jaeger API endpoint - such as started by `docker-compose up`.

Each daemon reports trace data independently, defaulting to not tracing.
Pass `-trace http://JAEGER_HOST:14268` to a daemon to enable tracing. `docker-compose.yml` has this
already configured.

By default traces are sampled at low frequency by each service. The `cachecash-curl` program will
force a trace when tracing is enabled - again by passing `-trace http://JAEGER_HOST:14268` to it;
for the docker-compose test network, the host to provide is localhost.

For that test network traces can be seen in the Jaeger UI at `http://localhost://16686`.

### Adding information

Any function that should show up in profiles can have a `Span` created in it:

    ```golang
    ctx, span := trace.StartSpan(ctx, "cachecash.com/Publisher/GetContent")
    defer span.End()
    ```

Spans can have additional tags and metadata attached to them:

    ```golang
    resp, err := pc.grpcClient.GetContent(ctx, req)
    if err != nil {
        span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
    }
    ```

The OpenCensus API documentation should be consulted for more details.

## Via client logs

Requires a running test network such as the one started by `docker-compose up`.

```
go build -o bin/cachecash-curl ./cmd/cachecash-curl && ./bin/cachecash-curl cachecash://localhost:7070/file0.bin 2>&1 | grep '^\{' | jq -c 'select(.kind=="timing") | del(.time, .level, .kind)'
```

This should produce output that looks like

```
{"msg":"getChunk","when":["2019-04-03T08:50:48.85573257-07:00","2019-04-03T08:50:48.857004708-07:00"]}
{"msg":"exchangeTicketL1","when":["2019-04-03T08:50:48.857045533-07:00","2019-04-03T08:50:48.857415144-07:00"]}
{"msg":"solvePuzzle","when":["2019-04-03T08:50:48.857755336-07:00","2019-04-03T08:50:48.858876709-07:00"]}
{"msg":"exchangeTicketL2","when":["2019-04-03T08:50:48.858912964-07:00","2019-04-03T08:50:48.859282353-07:00"]}
{"msg":"decryptData","when":["2019-04-03T08:50:48.859305972-07:00","2019-04-03T08:50:48.859496154-07:00"]}
```

You will need to pipe it through `jq -cs` to convert it into a single list before the visualizer will be able to load
it.
