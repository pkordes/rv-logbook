//go:build integration

package apitest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/apitest"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// tripRequest mirrors the CreateTripRequest / UpdateTripRequest JSON shape
// from openapi.yaml. Using a local struct keeps these tests self-contained
// without importing the generated handler types.
type tripRequest struct {
	Name      string  `json:"name"`
	StartDate string  `json:"start_date"` // YYYY-MM-DD
	EndDate   *string `json:"end_date,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

// tripResponse mirrors the Trip schema from openapi.yaml.
type tripResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	StartDate string  `json:"start_date"`
	EndDate   *string `json:"end_date,omitempty"`
	Notes     string  `json:"notes,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// tripListResponse mirrors the TripList schema from openapi.yaml.
type tripListResponse struct {
	Data       []tripResponse `json:"data"`
	Pagination struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination"`
}

// createTrip is a shared helper: POST /trips and register a DELETE cleanup.
// Returns the decoded trip so callers can use the ID for further operations.
func createTrip(t *testing.T, baseURL string, body tripRequest) tripResponse {
	t.Helper()

	b, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/trips", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "POST /trips should return 201")

	var trip tripResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&trip))
	require.NotEmpty(t, trip.ID, "created trip must have an ID")

	// Register cleanup so the test DB stays tidy even on failure.
	t.Cleanup(func() {
		req, _ := http.NewRequest(http.MethodDelete, baseURL+"/trips/"+trip.ID, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	})

	return trip
}

// TestTrip_Create verifies that POST /trips returns 201 and the created trip.
func TestTrip_Create(t *testing.T) {
	pool := testutil.NewPool(t)
	srv := apitest.NewServer(t, pool)

	trip := createTrip(t, srv.URL, tripRequest{
		Name:      "Summer Tour",
		StartDate: "2025-06-01",
		Notes:     "Pacific coast route",
	})

	assert.Equal(t, "Summer Tour", trip.Name)
	assert.Equal(t, "2025-06-01", trip.StartDate)
	assert.Equal(t, "Pacific coast route", trip.Notes)
	assert.NotEmpty(t, trip.CreatedAt)
}

// TestTrip_GetByID verifies that GET /trips/{id} returns the created trip.
func TestTrip_GetByID(t *testing.T) {
	pool := testutil.NewPool(t)
	srv := apitest.NewServer(t, pool)

	created := createTrip(t, srv.URL, tripRequest{
		Name:      "GetByID Test Trip",
		StartDate: "2025-07-01",
	})

	resp, err := http.Get(fmt.Sprintf("%s/trips/%s", srv.URL, created.ID))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got tripResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Name, got.Name)
}

// TestTrip_GetByID_NotFound verifies that GET /trips/{id} returns 404 for an unknown ID.
func TestTrip_GetByID_NotFound(t *testing.T) {
	pool := testutil.NewPool(t)
	srv := apitest.NewServer(t, pool)

	resp, err := http.Get(srv.URL + "/trips/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestTrip_List verifies that GET /trips returns a list containing the created trip.
func TestTrip_List(t *testing.T) {
	pool := testutil.NewPool(t)
	srv := apitest.NewServer(t, pool)

	// Use a name unique enough to find in the list even if other test data exists.
	uniqueName := fmt.Sprintf("ListTest-%d", time.Now().UnixNano())
	created := createTrip(t, srv.URL, tripRequest{
		Name:      uniqueName,
		StartDate: "2025-08-01",
	})

	resp, err := http.Get(srv.URL + "/trips")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var list tripListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))

	var found bool
	for _, item := range list.Data {
		if item.ID == created.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created trip %s should appear in GET /trips", created.ID)
}

// TestTrip_Update verifies that PATCH /trips/{id} updates and returns the trip.
func TestTrip_Update(t *testing.T) {
	pool := testutil.NewPool(t)
	srv := apitest.NewServer(t, pool)

	created := createTrip(t, srv.URL, tripRequest{
		Name:      "Before Update",
		StartDate: "2025-09-01",
	})

	end := "2025-09-30"
	update := tripRequest{
		Name:      "After Update",
		StartDate: "2025-09-01",
		EndDate:   &end,
		Notes:     "Updated notes",
	}
	b, err := json.Marshal(update)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/trips/%s", srv.URL, created.ID), bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got tripResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, "After Update", got.Name)
	assert.Equal(t, "Updated notes", got.Notes)
	require.NotNil(t, got.EndDate)
	assert.Equal(t, "2025-09-30", *got.EndDate)
}

// TestTrip_Delete verifies that DELETE /trips/{id} returns 204 and the trip is gone.
func TestTrip_Delete(t *testing.T) {
	pool := testutil.NewPool(t)
	srv := apitest.NewServer(t, pool)

	created := createTrip(t, srv.URL, tripRequest{
		Name:      "To Be Deleted",
		StartDate: "2025-10-01",
	})

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/trips/%s", srv.URL, created.ID), nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Confirm it really is gone.
	getResp, err := http.Get(fmt.Sprintf("%s/trips/%s", srv.URL, created.ID))
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}
