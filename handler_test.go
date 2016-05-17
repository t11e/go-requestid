package requestid_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/xhandler"
	"github.com/stretchr/testify/assert"
	"github.com/t11e/go-ctxlog"
	"github.com/t11e/go-requestid"
	"golang.org/x/net/context"
)

func TestFromContext(t *testing.T) {
	_, ok := requestid.FromContext(context.Background())
	assert.False(t, ok)

	ctx := requestid.NewContext(context.Background(), "test")
	actual, ok := requestid.FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "test", actual)
}

func testConfig(id string) requestid.Config {
	return requestid.Config{Generator: func() (string, error) {
		return id, nil
	}}
}

func testConfigErr(err error) requestid.Config {
	return requestid.Config{Generator: func() (string, error) {
		return "", err
	}}
}

func TestHeaderMiddleware(t *testing.T) {
	for idx, scenario := range []struct {
		request      *http.Request
		ctx          context.Context
		middleware   requestid.HeaderMiddleware
		expected     []byte
		expectedCode int
	}{
		{
			request: testRequest(t, "GET", "/", nil),
			ctx:     context.Background(),
			middleware: requestid.HeaderMiddleware{
				Config: testConfig("from-make"),
			},
			expected:     []byte("from-make"),
			expectedCode: http.StatusOK,
		},
		{
			request: testRequest(t, "GET", "/", nil),
			ctx:     context.Background(),
			middleware: requestid.HeaderMiddleware{
				Config: testConfigErr(errors.New("make error")),
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			request: testRequest(t, "GET", "/", nil),
			ctx:     context.Background(),
			middleware: requestid.HeaderMiddleware{
				Config:        testConfig("from-make"),
				RequestHeader: "Request-Id",
			},
			expected:     []byte("from-make"),
			expectedCode: http.StatusOK,
		},
		{
			request: testRequest(t, "GET", "/", nil),
			ctx:     context.Background(),
			middleware: requestid.HeaderMiddleware{
				Config:         testConfig("from-make"),
				ResponseHeader: "Request-Id",
			},
			expected:     []byte("from-make"),
			expectedCode: http.StatusOK,
		},
		{
			request: WithHeaders(testRequest(t, "GET", "/", nil), map[string][]string{"Request-Id": {"from-header"}}),
			ctx:     context.Background(),
			middleware: requestid.HeaderMiddleware{
				Config:        testConfig("from-make"),
				RequestHeader: "Request-Id",
			},
			expected:     []byte("from-header"),
			expectedCode: http.StatusOK,
		},
		{
			request: WithHeaders(testRequest(t, "GET", "/", nil), map[string][]string{"X-Input": {"from-header"}}),
			ctx:     context.Background(),
			middleware: requestid.HeaderMiddleware{
				Config:         testConfig("from-make"),
				RequestHeader:  "X-Input",
				ResponseHeader: "X-Output",
			},
			expected:     []byte("from-header"),
			expectedCode: http.StatusOK,
		},
		{
			request: WithHeaders(testRequest(t, "GET", "/", nil), map[string][]string{"Request-Id": {"from-header"}}),
			ctx:     context.Background(),
			middleware: requestid.HeaderMiddleware{
				Config:         testConfig("from-make"),
				RequestHeader:  "Request-Id",
				ResponseHeader: "Request-Id",
			},
			expected:     []byte("from-header"),
			expectedCode: http.StatusOK,
		},
		{
			request: testRequest(t, "GET", "/", nil),
			ctx:     requestid.NewContext(context.Background(), "from-ctx"),
			middleware: requestid.HeaderMiddleware{
				Config: testConfig("from-make"),
			},
			expected:     []byte("from-ctx"),
			expectedCode: http.StatusOK,
		},
		{
			request: WithHeaders(testRequest(t, "GET", "/", nil), map[string][]string{"Request-Id": {"from-header"}}),
			ctx:     requestid.NewContext(context.Background(), "from-ctx"),
			middleware: requestid.HeaderMiddleware{
				Config:         testConfig("from-make"),
				RequestHeader:  "Request-Id",
				ResponseHeader: "Request-Id",
			},
			expected:     []byte("from-ctx"),
			expectedCode: http.StatusOK,
		},
		{
			request:      WithHeaders(testRequest(t, "GET", "/", nil), map[string][]string{requestid.DefaultHeader: {"from-header"}}),
			ctx:          context.Background(),
			middleware:   requestid.DefaultHeaderMiddleware,
			expected:     []byte("from-header"),
			expectedCode: http.StatusOK,
		},
	} {
		var actual *string
		next := func(ctx context.Context, response http.ResponseWriter, request *http.Request) {
			if a, ok := requestid.FromContext(ctx); ok {
				actual = &a
			}
			response.WriteHeader(http.StatusOK)
		}
		response := httptest.NewRecorder()
		scenario.middleware.HandlerC(xhandler.HandlerFuncC(next)).ServeHTTPC(scenario.ctx, response, scenario.request)
		if actual == nil {
			assert.Nil(t, scenario.expected, "Scenario %d", idx)
		} else {
			if scenario.expected != nil {
				expected := string(scenario.expected)
				if expected != *actual {
					t.Errorf("Scenario %d\nexpected: %s\nactual: %s\n", idx, expected, *actual)
				}
			}
			if scenario.middleware.ResponseHeader != "" {
				assert.Equal(t, *actual, response.Header().Get(scenario.middleware.ResponseHeader))
			}
		}
		if scenario.expectedCode != response.Code {
			t.Errorf("Scenario %d, expected=%d actual=%d\n", idx, scenario.expectedCode, response.Code)
		}
	}
}

func TestLoggerMiddleware(t *testing.T) {
	buf := bytes.Buffer{}
	response := httptest.NewRecorder()
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctxlog.FromContext(ctx).Info("hello from test")
	}
	requestid.LoggerMiddleware{&buf}.HandlerC(xhandler.HandlerFuncC(next)).ServeHTTPC(requestid.NewContext(context.Background(), "testreqid"), response, nil)
	assert.Equal(t, "[testreqid] hello from test\n", buf.String())
}

func TestLoggerMiddleware_MissingRequestID(t *testing.T) {
	buf := bytes.Buffer{}
	response := httptest.NewRecorder()
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctxlog.FromContext(ctx).Info("hello from test")
	}
	requestid.LoggerMiddleware{&buf}.HandlerC(xhandler.HandlerFuncC(next)).ServeHTTPC(context.Background(), response, nil)
	assert.Equal(t, "hello from test\n", buf.String())
}

func testRequest(t *testing.T, method string, url string, body io.Reader) *http.Request {
	if body == nil {
		body = bytes.NewBuffer([]byte{})
	}
	r, err := http.NewRequest(method, url, body)
	assert.NoError(t, err)
	return r
}

func WithHeaders(r *http.Request, h map[string][]string) *http.Request {
	for k, v := range h {
		r.Header[k] = v
	}
	return r
}
