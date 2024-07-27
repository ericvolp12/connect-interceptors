# ConnectRPC Interceptors
A library of interceptors for ConnectRPC that I've found handy

## Interceptors

- `pkg/retry` - An interceptor that allows you to retry RPCs that fail with specific error conditions.
    - Supports custom backoff strategies and custom functions for parsing errors to determine if you should retry.
    - Defaults to a 1 second backoff, 10 retry maximum, and automatic retries on connection errors.

## Using an Interceptor

Read more about ConnectRPC Interceptors [here](https://connectrpc.com/docs/go/interceptors)

To plug any of these interceptors into your ConnectRPC Client or Server, you can initialize your server or client as follows:

```go
import (
    "github.com/ericvolp12/connect-interceptors/pkg/retry"
)

interceptors := connect.WithInterceptors(retry.NewRetryInterceptor(nil, nil))

// For handlers:
mux := http.NewServeMux()
mux.Handle(greetv1connect.NewGreetServiceHandler(
  &greetServer{},
  interceptors,
))

// For clients:
client := greetv1connect.NewGreetServiceClient(
  http.DefaultClient,
  "https://api.acme.com",
  interceptors,
)
```
