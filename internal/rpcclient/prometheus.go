package rpcclient

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var totalRpcCallsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "app_rpc_calls_total",
	Help: "The total number of RPC calls made.",
}, []string{"request", "result"})

var rpcCallDurations = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "app_rpc_call_duration_seconds",
	Help:    "RPC calls latencies in seconds.",
	Buckets: prometheus.DefBuckets,
}, []string{"request"})

type prometheusClient struct {
	rpcClient RpcClient
}

func NewPrometheusClient(rpcClient RpcClient) RpcClient {
	return &prometheusClient{rpcClient: rpcClient}
}

func (c *prometheusClient) Call(ctx context.Context, req Message) (Message, error) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		rpcCallDurations.With(map[string]string{
			"request": req.Name,
		}).Observe(v)
	}))
	defer timer.ObserveDuration()

	resp, err := c.rpcClient.Call(ctx, req)
	if err != nil {
		totalRpcCallsCounter.With(map[string]string{"request": req.Name, "result": "error"}).Inc()

		return Message{}, err
	}

	totalRpcCallsCounter.With(map[string]string{"request": req.Name, "result": "success"}).Inc()

	return resp, nil
}

func (c *prometheusClient) Close() error {
	return c.rpcClient.Close()
}
