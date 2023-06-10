package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	internalJSON "github.com/jbrekelmans/go-graphql/json"
)

// Client is a client for talking to a GraphQL server using HTTP with JSON-encoded requests/responses.
type Client struct {
	url        string
	httpClient *http.Client
}

// NewClient constructs a client.
func NewClient(url string, httpClient *http.Client) *Client {
	c := &Client{
		url:        url,
		httpClient: httpClient,
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c
}

func (c *Client) doRequest(ctx context.Context, operationType string, q any, variables map[string]any) (*http.Response, error) {
	var queryBuilder queryBuilder
	if err := queryBuilder.operation(operationType, q, variables); err != nil {
		return nil, err
	}
	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(request{
		Query:     queryBuilder.String(),
		Variables: variables,
	}); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, &reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	respBodyBytes, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return resp, fmt.Errorf(`error reading body of %d-response: %w`, resp.StatusCode, err)
	}
	var respBody response
	if err := json.NewDecoder(bytes.NewReader(respBodyBytes)).Decode(&respBody); err != nil {
		return resp, fmt.Errorf(`error unmarshalling body of %d-response: %w`, resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		err := NewErrorf(`response has non-success status %d: %s`, resp.StatusCode, string(respBodyBytes))
		err.Errors = respBody.Errors
		return resp, err
	}
	if respBody.Data != nil {
		if err := internalJSON.Unmarshal(*respBody.Data, q); err != nil {
			err := NewErrorf(`error decoding data of %d-response: %w`, resp.StatusCode, err)
			err.Errors = respBody.Errors
			return resp, err
		}
	}
	if len(respBody.Errors) > 0 {
		errorsJSON, _ := json.Marshal(respBody.Errors)
		err := NewErrorf(`%d-response with errors: %s`, resp.StatusCode, string(errorsJSON))
		err.Errors = respBody.Errors
		return resp, err
	}
	return resp, nil
}

// Mutate does a mutation operation on the GraphQL server.
// See Query for more information.
func (c *Client) Mutate(ctx context.Context, m any, variables map[string]any) (*http.Response, error) {
	return c.doRequest(ctx, "mutation", m, variables)
}

// Query does a query operation on the GraphQL server.
// If the HTTP response status and headers were received successfully then returns a non-nil *http.Response that reflects the status and
// headers. The body of the returned HTTP response is always closed.
//
// If the GraphQL response was completely received and parsed, and contains GraphQL-level errors, an error of type *Error is returned
// that reflects the GraphQL-level errors.
//
// Users should interpret any of the following as a transient error condition that may go away by retrying:
// 1. The returned error, say x, implements interface{ Temporary() bool } and x.Temporary() is true.
// 2. The returned error wraps an error, say x, that implements interface{ Temporary() bool } and x.Temporary() is true.
// 3. The returned error, say x, implements interface{ Timeout() bool } and x.Timeout() is true.
// 4. The returned error wraps an error, say x, that implements interface{ Timeout() bool } and x.Timeout() is true.
// 5. The returned response has status code between 500 and 599.
// 6. The returned response has status code 429.
//
// 1, 2, 3 and 4 are known to be produced in the event of transient errors and timeouts by
//   - the *http.Client;
//   - the http.RoundTripper of the *http.Client when dialing, sending the HTTP request and reading
//     the HTTP response; -and
//   - the underlying connnection when reading the HTTP response body.
//
// See https://spec.graphql.org/.
func (c *Client) Query(ctx context.Context, q any, variables map[string]any) (*http.Response, error) {
	return c.doRequest(ctx, "query", q, variables)
}

type request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type response struct {
	Data   *json.RawMessage `json:"data"`
	Errors []ErrorItem      `json:"errors"`
}
