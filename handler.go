package requestid

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/xhandler"
	"github.com/t11e/go-ctxlog"
	"golang.org/x/net/context"
)

type contextKey int

const (
	requestIDKey contextKey = iota
)

func FromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(requestIDKey).(string)
	return v, ok
}

func NewContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

const DefaultHeader = "Request-Id"

var DefaultHeaderMiddleware = HeaderMiddleware{RequestHeader: DefaultHeader, ResponseHeader: DefaultHeader}

type HeaderMiddleware struct {
	Config
	RequestHeader  string
	ResponseHeader string
}

func (m HeaderMiddleware) HandlerC(next xhandler.HandlerC) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		id, ok := FromContext(ctx)
		if !ok {
			if m.RequestHeader != "" {
				id = r.Header.Get(m.RequestHeader)
			}
			if id == "" {
				var err error
				id, err = m.Config.MakeID()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			ctx = NewContext(ctx, id)
		}
		if m.ResponseHeader != "" {
			w.Header().Set(m.ResponseHeader, id)
		}
		next.ServeHTTPC(ctx, w, r)
	})
}

type LoggerMiddleware struct {
	io.Writer
}

func (m LoggerMiddleware) HandlerC(next xhandler.HandlerC) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		prefix := ""
		if id, ok := FromContext(ctx); ok {
			prefix = fmt.Sprintf("[%s] ", id)
		}
		config := ctxlog.Direct{Writer: m.Writer, Prefix: prefix}
		ctx = ctxlog.NewContext(ctx, config.Logger("context"))
		next.ServeHTTPC(ctx, w, r)
	})
}
