package dbtx

import (
	"context"
	"database/sql"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// DBInjector provides context injectors
type DBInjector struct {
	db *sql.DB
}

func NewInjector(db *sql.DB) *DBInjector {
	return &DBInjector{db: db}
}

func (i *DBInjector) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = ContextWithExecutor(ctx, i.db)
		return handler(ctx, req)
	}
}

func (i *DBInjector) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		stream := grpc_middleware.WrapServerStream(ss)
		stream.WrappedContext = ContextWithExecutor(stream.Context(), i.db)
		return handler(srv, stream)
	}
}
