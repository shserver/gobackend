package middleware

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	ctxTimeout = 500 * time.Millisecond
)

func UnaryClient() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		log.Printf("unary interceptor method : %s", method)
		ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryServer() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "No metadata")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "No authorization token")
		}

		accessToken := values[0]
		_, err = VerifyJWT(accessToken)
		if err != nil {
			log.Printf("Auth Fail : %v", err)
			return nil, status.Errorf(codes.Unauthenticated, "Invalid authorization token")
		}
		return handler(ctx, req)
	}
}
