package main

import (
	"fmt"
	"github.com/Snakdy/lambda-function-url/pkg/invoke"
	"github.com/cenkalti/backoff/v4"
	"github.com/kelseyhightower/envconfig"
	"github.com/lpar/problem"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type environment struct {
	Port     int    `envconfig:"PORT" default:"8080"`
	Upstream string `split_words:"true" required:"true"`
	Timeout  int64  `split_words:"true" default:"30"`
}

func main() {
	var e environment
	envconfig.MustProcess("app", &e)

	var client *rpc.Client
	var err error

	// connect to the upstream function.
	// We might start before it, so we need retry-backoff
	// logic
	err = backoff.Retry(func() error {
		slog.Info("connecting to function", "address", e.Upstream)
		client, err = rpc.Dial("tcp", e.Upstream)
		if err != nil {
			slog.Error("failed to connect to upstream", "error", err)
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff(backoff.WithInitialInterval(time.Second)))
	if err != nil {
		slog.Error("could not connect to upstream after all retries")
		os.Exit(1)
	}

	svc := invoke.NewService(client, e.Timeout)

	r := http.NewServeMux()
	r.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
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
