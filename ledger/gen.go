package ledger

//go:generate protoc --gofast_out=plugins=grpc:. --proto_path=.:../vendor:../ccmsg block.proto
