package invoke

import "net/rpc"

type RequestContext struct {
	TraceID   string
	RequestID string
}

type Service struct {
	rpcClient *rpc.Client
	timeout   int64
}
