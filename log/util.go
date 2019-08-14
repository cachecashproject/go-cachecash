package log

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcFailedError(err error) error {
	return status.Errorf(codes.FailedPrecondition, "%v", err)
}
