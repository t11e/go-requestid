# go-requestid

[![Build Status](https://semaphoreci.com/api/v1/projects/bc0f95e5-c45d-45e7-b5b2-f3e868d01793/814693/badge.svg)](https://semaphoreci.com/t11e/go-requestid)

Context aware unique request ID management.

The request ID can serve as a useful tracer bullet when diagnosing issues in a distributed system.

The `DefaultHeaderMiddleware` will use the `Request-Id` HTTP request header to populate the ID or automatically
generate one if it is missing. The response will contain the ID in the `Request-Id` HTTP response header. Any outgoing
connections made by your application should propagate this ID for best diagnostic support.

## Usage

Add this library as a dependency to your project.

```bash
glide get github.com/t11e/go-requestid
```

To make the request ID available add the `DefaultHeaderMiddleware` to your application.

```go
chain.UseC(requestid.DefaultHeaderMiddleware.HandlerC)
```

To obtain a logger bound to the current request ID use the `LoggerMiddleware`.

```go
chain.UseC(requestid.LoggerMiddleware{os.Stdout}.HandlerC)
```

In each handler that needs a request ID:

```go
func handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
        id, ok := requestid.FromContext(ctx)
        // ...
}
```

## Development

```bash
brew install go glide
glide install
go test $(go list ./... | grep -v /vendor/)
```

Be sure to install our standard [go commit hook](https://github.com/t11e/development-environment#golang-checks).

To run goimports without messing up `vendor/` use `goimports -w $(find . -name '*.go' | grep -v /vendor/)`.
