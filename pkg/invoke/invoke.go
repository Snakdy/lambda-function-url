package invoke

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"golang.org/x/exp/slog"
	"net/rpc"
	"time"
)

func NewService(rpcClient *rpc.Client, timeout int64) *Service {
	return &Service{
		rpcClient: rpcClient,
		timeout:   timeout,
	}
}

func (svc *Service) Invoke(request any, response any, requestCtx RequestContext) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("converting request to json: %w", err)
	}
	
	res, err := svc.InvokeJSON(payload, requestCtx)
	if err != nil {
		return err
	}
	// read the response back into
	// the callers struct
	if err := json.Unmarshal(res, &response); err != nil {
		return fmt.Errorf("reading response json: %w", err)
	}
	slog.Info("successfully invoked function")
	return nil
}

func (svc *Service) InvokeJSON(request []byte, requestCtx RequestContext) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("converting request to json: %w", err)
	}
	// assemble the request
	req := messages.InvokeRequest{
		Payload: payload,
		Deadline: messages.InvokeRequest_Timestamp{
			Seconds: time.Now().Unix() + svc.timeout,
		},
		RequestId:    requestCtx.RequestID,
		XAmznTraceId: requestCtx.TraceID,
	}

	slog.Info("invoking function", "requestId", requestCtx.RequestID, "traceId", requestCtx.TraceID)
	var res messages.InvokeResponse
	// make the call
	err = svc.rpcClient.Call("Function.Invoke", req, &res)
	if err != nil {
		return nil, fmt.Errorf("invoking function: %w", err)
	}
	if res.Error != nil {
		return nil, fmt.Errorf("function returned error: %v", res.Error)
	}

	slog.Info("successfully invoked function")
	return res.Payload, nil
}
