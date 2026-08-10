package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

func clientInterceptor(
	ctx context.Context, method string,
	req interface{}, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {

	return invoker(ctx, method, req, reply, cc, opts...)
}
