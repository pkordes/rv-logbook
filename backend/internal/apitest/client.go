//go:build integration

package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TripRequest is the input shape for create and update trip operations.
// Maps to CreateTripRequest / UpdateTripRequest in openapi.yaml.
type TripRequest struct {
	Name      string  `json:"name"`
	StartDate string  `json:"start_date"` // YYYY-MM-DD
	EndDate   *string `json:"end_date,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

// TripResponse is the decoded response for a single trip.
// Maps to the Trip schema in openapi.yaml.
type TripResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	StartDate string  `json:"start_date"`
	EndDate   *string `json:"end_date,omitempty"`
	Notes     string  `json:"notes,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// TripListResponse is the decoded response for the trip list endpoint.
// Maps to the TripList schema in openapi.yaml.
type TripListResponse struct {
	Data       []TripResponse `json:"data"`
	Pagination struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination"`
}

// Client is a typed HTTP client for the RV Logbook API.
// It wraps net/http to remove boilerplate from test files — tests
// express intent rather than HTTP mechanics.
//
// All methods call t.Fatal on unexpected errors so test bodies stay linear
// (no err != nil checks). Status-code failures are surfaced as clear
// require messages rather than panics.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient returns a Client pointed at baseURL.
// Pair it with NewServer to get a fully-wired test client in two lines:
//
//	srv := apitest.NewServer(t, pool)
//	c   := apitest.NewClient(srv.URL)
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: http.DefaultClient}
}

// CreateTrip POSTs to /trips and returns the created trip.
// Registers a DELETE cleanup so the row is removed when the test finishes,
// even if the test fails partway through.
// Fails the test immediately if the response status is not 201.
func (c *Client) CreateTrip(t *testing.T, req TripRequest) TripResponse {
	t.Helper()
	var trip TripResponse
	c.do(t, http.MethodPost, "/trips", req, http.StatusCreated, &trip)
	require.NotEmpty(t, trip.ID, "created trip must have an ID")
	t.Cleanup(func() {
		r, _ := http.NewRequest(http.MethodDelete, c.baseURL+"/trips/"+trip.ID, nil)
		resp, err := c.http.Do(r)
		if err == nil {
			resp.Body.Close()
		}
	})
	return trip
}

// GetTrip GETs /trips/{id} and returns the trip.
// Fails the test immediately if the response status is not 200.
func (c *Client) GetTrip(t *testing.T, id string) TripResponse {
	t.Helper()
	var trip TripResponse
	c.do(t, http.MethodGet, "/trips/"+id, nil, http.StatusOK, &trip)
	return trip
}

// GetTripStatus GETs /trips/{id} and returns only the HTTP status code.
// Use this when you want to assert on a non-200 response (e.g. 404).
func (c *Client) GetTripStatus(t *testing.T, id string) int {
	t.Helper()
	return c.status(t, http.MethodGet, "/trips/"+id, nil)
}

// ListTrips GETs /trips and returns the paginated list.
// Fails the test immediately if the response status is not 200.
func (c *Client) ListTrips(t *testing.T) TripListResponse {
	t.Helper()
	var list TripListResponse
	c.do(t, http.MethodGet, "/trips", nil, http.StatusOK, &list)
	return list
}

// UpdateTrip PUTs to /trips/{id} and returns the updated trip.
// Fails the test immediately if the response status is not 200.
func (c *Client) UpdateTrip(t *testing.T, id string, req TripRequest) TripResponse {
	t.Helper()
	var trip TripResponse
	c.do(t, http.MethodPut, "/trips/"+id, req, http.StatusOK, &trip)
	return trip
}

// DeleteTrip sends DELETE /trips/{id}.
// Fails the test immediately if the response status is not 204.
func (c *Client) DeleteTrip(t *testing.T, id string) {
	t.Helper()
	c.do(t, http.MethodDelete, "/trips/"+id, nil, http.StatusNoContent, nil)
}

// -------------------------------------------------------------------------
// Internal helpers — keep test method bodies free of HTTP plumbing.
// -------------------------------------------------------------------------

// do sends a request with an optional JSON body, asserts the expected status,
// and decodes the response body into out (if out is non-nil).
func (c *Client) do(t *testing.T, method, path string, body any, wantStatus int, out any) {
	t.Helper()
	resp := c.raw(t, method, path, body)
	defer resp.Body.Close()
	require.Equal(t, wantStatus, resp.StatusCode,
		"%s %s: unexpected status", method, path)
	if out != nil {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(out),
			"decode response from %s %s", method, path)
	}
}

// status sends a request and returns the HTTP status code without asserting.
func (c *Client) status(t *testing.T, method, path string, body any) int {
	t.Helper()
	resp := c.raw(t, method, path, body)
	resp.Body.Close()
	return resp.StatusCode
}

// raw builds and executes the HTTP request, failing the test on transport errors.
func (c *Client) raw(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err, "marshal request body for %s %s", method, path)
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.baseURL, path), bodyReader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	require.NoError(t, err, "execute %s %s", method, path)
	return resp
}
