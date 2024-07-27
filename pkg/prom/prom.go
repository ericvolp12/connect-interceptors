package prom

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

func NewMetricsInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			service, method := serviceAndMethod(req.Spec().Procedure)
			streamType := req.Spec().StreamType.String()

			// Measure Request Size
			if req != nil {
				if msg, ok := req.Any().(proto.Message); ok {
					TotalRequestBytesCounter.WithLabelValues(streamType, service, method).Add(float64(proto.Size(msg)))
				}
			}

			// Request started
			start := time.Now()
			TotalRequestsStartedCounter.WithLabelValues(streamType, service, method).Inc()

			// Pass to next handler
			res, err := next(ctx, req)
			elapsedSeconds := time.Since(start).Seconds()
			code := codeText(err)

			// Request completed
			TotalRequestsCompletedCounter.WithLabelValues(streamType, service, method, code).Inc()
			RequestLatencySecondsHistogram.WithLabelValues(streamType, service, method, code).Observe(elapsedSeconds)

			// Measure response size
			if err == nil {
				if msg, ok := res.Any().(proto.Message); ok {
					TotalResponseBytesCounter.WithLabelValues(streamType, service, method).Add(float64(proto.Size(msg)))
				}
			}

			return res, err
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

// serviceAndMethod returns the service and method from a procedure.
func serviceAndMethod(procedure string) (string, string) {
	procedure = strings.TrimPrefix(procedure, "/")
	service, method := "unknown", "unknown"
	if strings.Contains(procedure, "/") {
		long := strings.Split(procedure, "/")[0]
		if strings.Contains(long, ".") {
			service = strings.Split(long, ".")[0]
		}
	}
	if strings.Contains(procedure, "/") {
		method = strings.Split(procedure, "/")[1]
	}
	return service, method
}

// codeText returns the code text name for an error.
func codeText(err error) string {
	if err == nil {
		return "success"
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		return "unknown"
	}
	return connectErr.Code().String()
}
