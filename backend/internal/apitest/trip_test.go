//go:build integration

package apitest_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkordes/rv-logbook/backend/internal/apitest"
	"github.com/pkordes/rv-logbook/backend/testutil"
)

// newClient is a one-liner to get a fully-wired test client.
// Every test calls this to avoid repeating the server + pool setup.
func newClient(t *testing.T) *apitest.Client {
	t.Helper()
	pool := testutil.NewPool(t)
	srv := apitest.NewServer(t, pool)
	return apitest.NewClient(srv.URL)
}

// TestTrip_Create verifies that POST /trips returns the created trip.
func TestTrip_Create(t *testing.T) {
	c := newClient(t)

	trip := c.CreateTrip(t, apitest.TripRequest{
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
	c := newClient(t)

	created := c.CreateTrip(t, apitest.TripRequest{Name: "GetByID Test", StartDate: "2025-07-01"})
	got := c.GetTrip(t, created.ID)

	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Name, got.Name)
}

// TestTrip_GetByID_NotFound verifies that GET /trips/{id} returns 404 for an unknown ID.
func TestTrip_GetByID_NotFound(t *testing.T) {
	c := newClient(t)

	assert.Equal(t, http.StatusNotFound, c.GetTripStatus(t, "00000000-0000-0000-0000-000000000000"))
}

// TestTrip_List verifies that GET /trips includes the created trip.
func TestTrip_List(t *testing.T) {
	c := newClient(t)

	// Unique name so we can find this trip even if other data exists.
	uniqueName := fmt.Sprintf("ListTest-%d", time.Now().UnixNano())
	created := c.CreateTrip(t, apitest.TripRequest{Name: uniqueName, StartDate: "2025-08-01"})

	list := c.ListTrips(t)

	var found bool
	for _, item := range list.Data {
		if item.ID == created.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created trip should appear in GET /trips")
}

// TestTrip_Update verifies that PUT /trips/{id} updates and returns the trip.
func TestTrip_Update(t *testing.T) {
	c := newClient(t)

	created := c.CreateTrip(t, apitest.TripRequest{Name: "Before Update", StartDate: "2025-09-01"})

	end := "2025-09-30"
	updated := c.UpdateTrip(t, created.ID, apitest.TripRequest{
		Name:      "After Update",
		StartDate: "2025-09-01",
		EndDate:   &end,
		Notes:     "Updated notes",
	})

	assert.Equal(t, "After Update", updated.Name)
	assert.Equal(t, "Updated notes", updated.Notes)
	require.NotNil(t, updated.EndDate)
	assert.Equal(t, "2025-09-30", *updated.EndDate)
}

// TestTrip_Delete verifies that DELETE /trips/{id} removes the trip.
func TestTrip_Delete(t *testing.T) {
	c := newClient(t)

	created := c.CreateTrip(t, apitest.TripRequest{Name: "To Be Deleted", StartDate: "2025-10-01"})

	c.DeleteTrip(t, created.ID)

	assert.Equal(t, http.StatusNotFound, c.GetTripStatus(t, created.ID))
}
