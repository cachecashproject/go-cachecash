package ccmsg

//go:generate protoc --gofast_out=plugins=grpc:. --proto_path=.:../vendor:.. common.proto client_publisher.proto client_cache.proto bootstrap.proto cache_publisher.proto ledger.proto publisher_cache.proto faucet.proto
