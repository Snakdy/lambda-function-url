package main

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/lpar/problem"
	"github.com/snakdy/lambda-function-url/pkg/invoke"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
	"net/rpc"
	"os"
)

type environment struct {
	Port     int    `envconfig:"PORT" default:"8080"`
	Upstream string `required:"true"`
	Timeout int64 `default:"30"`
}

func main() {
	var e environment
	envconfig.MustProcess("app", &e)
	
	rpcClient, err := rpc.Dial("tcp", e.Upstream)
	if err != nil {
		slog.Error("failed to connect to upstream", "error", err)
		os.Exit(1)
	}

	svc := invoke.NewService(rpcClient, e.Timeout)
	
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			_ = problem.MustWrite(w, problem.New(http.StatusBadRequest).WithErr(err))
			return
		}
		defer r.Body.Close()
		resp, err := svc.InvokeJSON(body, invoke.RequestContext{
			TraceID:   r.Header.Get("X-Amzn-Trace-ID"),
			RequestID: r.Header.Get("X-Amzn-RequestID"),
		})
		if err != nil {
			_ = problem.MustWrite(w, problem.New(http.StatusInternalServerError).WithErr(err))
			return
		}
		_, _ = w.Write(resp)
	})
	
	slog.Error("server exited", "error", http.ListenAndServe(fmt.Sprintf(":%d", e.Port), r))
}
