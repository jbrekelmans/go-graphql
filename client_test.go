package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Client(t *testing.T) {
	setupTestCase := func(statusCode int, respBody []byte, respBodyReadErr error) *Client {
		transport := &testTransport{
			RespBody:        respBody,
			RespBodyReadErr: respBodyReadErr,
			StatusCode:      statusCode,
		}
		if respBody == nil {
			transport.Err = fmt.Errorf(`error doing request`)
		}
		httpClient := &http.Client{
			Transport: transport,
		}
		c := &Client{
			httpClient: httpClient,
			url:        "http://localhost/graphql",
		}
		return c
	}
	t.Run("doRequest", func(t *testing.T) {
		t.Run("InvalidQuery", func(t *testing.T) {
			c := setupTestCase(0, nil, nil)
			var q int
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.ErrorContains(t, err, "invalid query type *int")
		})
		t.Run("ErrorMarshalingRequestBody", func(t *testing.T) {
			c := setupTestCase(0, nil, nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, map[string]interface{}{
				"test": jsonMarshalBomb{},
			})
			assert.ErrorContains(t, err, "jsonMarshalBomb: boom!")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query($test:jsonMarshalBomb!){name}", err2.Operation)
			}
		})
		t.Run("ErrorNewRequest", func(t *testing.T) {
			c := setupTestCase(0, nil, nil)
			c.url = "\x00invalid-nul-byte-in-url"
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.ErrorContains(t, err, "invalid control character in URL")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
			}
		})
		t.Run("ErrorDoingRequest", func(t *testing.T) {
			c := setupTestCase(0, nil, nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.ErrorContains(t, err, "error doing request")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
			}
		})
		t.Run("ErrorReadingResponseBody", func(t *testing.T) {
			c := setupTestCase(0, []byte(`{`), errors.New("error reading body"))
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.ErrorContains(t, err, "error reading body")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
			}
		})
		t.Run("ErrorUnmarshalingResponse", func(t *testing.T) {
			c := setupTestCase(0, []byte(`{"errors":"this-should-be-an-array"}`), nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.ErrorContains(t, err, "error unmarshaling body of 200-response: ")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
			}
		})
		t.Run("UnexpectedStatusCode", func(t *testing.T) {
			c := setupTestCase(201, []byte(`{"errors":[{"message":"msg1","type":"type-entry-is-unspecified-in-graphql-spec"}]}`), nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.ErrorContains(t, err, "response has non-success status 201: ")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
				assert.Equal(t, []ErrorItem{
					{
						Message: "msg1",
						Raw: map[string]json.RawMessage{
							"message": json.RawMessage(`"msg1"`),
							"type":    json.RawMessage(`"type-entry-is-unspecified-in-graphql-spec"`),
						},
					},
				}, err2.Errors)
			}
		})
		t.Run("ErrorsAndPartialData", func(t *testing.T) {
			c := setupTestCase(200, []byte(`{"data":{"name":"name123"},"errors":[{"message":"msg2"}]}`), nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.Equal(t, "name123", q.Name)
			assert.ErrorContains(t, err, "200-response with errors: ")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
				assert.Equal(t, []ErrorItem{
					{
						Message: "msg2",
						Raw: map[string]json.RawMessage{
							"message": json.RawMessage(`"msg2"`),
						},
					},
				}, err2.Errors)
			}
		})
		t.Run("200WithErrors", func(t *testing.T) {
			c := setupTestCase(200, []byte(`{"errors":[{"message":"msg2"}]}`), nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.ErrorContains(t, err, "200-response with errors: ")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
				assert.Equal(t, []ErrorItem{
					{
						Message: "msg2",
						Raw: map[string]json.RawMessage{
							"message": json.RawMessage(`"msg2"`),
						},
					},
				}, err2.Errors)
			}
		})
		t.Run("ErrorUnmarshalingDataIntoReceiver", func(t *testing.T) {
			c := setupTestCase(200, []byte(`{"data":{"name":"hi"}}`), nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", q, nil)
			assert.ErrorContains(t, err, "error decoding data of 200-response: ")
			if assert.IsType(t, &Error{}, err) {
				err2 := err.(*Error)
				assert.Equal(t, "query{name}", err2.Operation)
			}
		})
		t.Run("Success", func(t *testing.T) {
			c := setupTestCase(200, []byte(`{"data":{"name":"hi"}}`), nil)
			var q struct {
				Name string
			}
			_, err := c.doRequest(context.Background(), "query", &q, nil)
			assert.NoError(t, err)
			assert.Equal(t, "hi", q.Name)
		})
	})
}

type jsonMarshalBomb struct {
}

var _ json.Marshaler = (*jsonMarshalBomb)(nil)

func (j jsonMarshalBomb) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf(`jsonMarshalBomb: boom!`)
}

type testTransport struct {
	Err             error
	RespBody        []byte
	RespBodyReadErr error
	StatusCode      int
}

var _ http.RoundTripper = (*testTransport)(nil)

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Err != nil {
		return nil, t.Err
	}
	if req.Method == http.MethodHead {
		return nil, fmt.Errorf(`HEAD method not supported`)
	}
	sc := t.StatusCode
	if sc == 0 {
		sc = 200
	}
	respHeader := http.Header{}
	resp := &http.Response{
		Status:     fmt.Sprintf("%d %s", sc, http.StatusText(sc)),
		StatusCode: sc,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Header:     respHeader,
	}
	respHeader.Set("Content-Type", "application/json")
	if req.Method != http.MethodHead {
		resp.Body = io.NopCloser(&testReader{
			Data: t.RespBody,
			Err:  t.RespBodyReadErr,
		})
		resp.ContentLength = int64(len(t.RespBody))
		respHeader.Set("Content-Length", fmt.Sprintf("%d", t.RespBody))
	}
	return resp, nil
}

type testReader struct {
	Data []byte
	Err  error
}

var _ io.Reader = (*testReader)(nil)

func (t *testReader) Read(p []byte) (n int, err error) {
	n = copy(p, t.Data)
	t.Data = t.Data[n:]
	if n == 0 {
		err = t.Err
		if err == nil {
			err = io.EOF
		}
	}
	return
}
