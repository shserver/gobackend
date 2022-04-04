package middleware

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type logInfo struct {
	request     string
	elapsedTime string
	response    codes.Code
}

const (
	ctxTimeout = 500 * time.Millisecond
)

var (
	authMethods = map[string]bool{
		"/sehyoung.pb.TestService/Hello": true,
	}
)

func UnaryClient() grpc.UnaryClientInterceptor {
	return grpc_middleware.ChainUnaryClient(
		UnaryClientTimeout(),
		UnaryClientLog(),
		// UnaryClientAuth(),
	)
}

func UnaryServer() grpc.ServerOption {
	return grpc_middleware.WithUnaryServerChain(
		UnaryServerAuth(),
	)
}

func UnaryClientTimeout() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryClientLog() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		startTime := time.Now()
		packagetest.InitLogger("")
		err := invoker(ctx, method, req, reply, cc, opts...)
		fiedls := &logInfo{
			request:     method[strings.LastIndex(method, ".")+1:],
			elapsedTime: time.Since(startTime).String(),
			response:    status.Code(err),
		}
		packagetest.LogFields("INFO", "Interceptor", fiedls)
		return err
	}
}

func UnaryClientAuth() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if authMethods[method] {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				return status.Errorf(codes.Unauthenticated, "No metadata")
			}

			values := md["authorization"]
			if len(values) == 0 {
				return status.Errorf(codes.Unauthenticated, "No authorization token")
			}

			accessToken := values[0]
			_, err := VerifyJWT(accessToken)
			if err != nil {
				return status.Errorf(codes.Unauthenticated, "Invalid authorization token")
			}

		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryServerAuth() grpc.UnaryServerInterceptor {
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
			return nil, status.Errorf(codes.Unauthenticated, "Invalid authorization token")
		}
		return handler(ctx, req)
	}
}

func UnaryServerLog() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return handler(ctx, req)
	}
}
