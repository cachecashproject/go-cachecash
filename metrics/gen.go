package metrics

//go:generate protoc --gofast_out=plugins=grpc:. --proto_path=.:../vendor metrics.proto
