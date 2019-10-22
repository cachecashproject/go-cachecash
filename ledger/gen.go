package ledger

//go:generate protoc --gofast_out=plugins=grpc:. --proto_path=.:../vendor block.proto
//go:generate find github.com/cachecashproject/go-cachecash/ledger/ -type f -exec mv -t . {} ;
//go:generate rm -r github.com
