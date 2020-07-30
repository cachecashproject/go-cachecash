package blockexplorer

//go:generate protoc --gofast_out=plugins=grpc:. --proto_path=.:../vendor:../ccmsg:.. blockexplorer.proto
//go:generate packr -v
