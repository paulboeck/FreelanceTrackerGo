package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupClientTest(t *testing.T) (*ClientHandlers, *testutil.TestDatabase, int) {
	testDB := testutil.SetupTestSQLite(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create handlers
	clientModel := models.NewClientModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	handlers := NewClientHandlers(clientModel, projectModel)

	return handlers, testDB, userID
}

func TestClientHandlers_ListClients(t *testing.T) {
	handlers, testDB, userID := setupClientTest(t)
	defer testDB.Cleanup(t)

	// Create some test clients
	testDB.InsertTestClient(t, "Client 1")
	testDB.InsertTestClient(t, "Client 2")
	testDB.InsertTestClient(t, "Client 3")

	t.Run("list all clients", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.ListClients(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		clients, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, clients, 3)

		// Verify metadata
		assert.NotNil(t, resp.Meta)
	})

	t.Run("list clients with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients?page=1&pageSize=2", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.ListClients(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		clients, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, clients, 2)

		// Verify pagination metadata
		assert.NotNil(t, resp.Meta)
		assert.Equal(t, 1, resp.Meta.Page)
		assert.Equal(t, 2, resp.Meta.PageSize)
		assert.Equal(t, 3, resp.Meta.Total)
	})

	t.Run("search clients", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients?search=Client+1", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.ListClients(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		clients, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, clients, 1)

		client := clients[0].(map[string]interface{})
		assert.Equal(t, "Client 1", client["name"])
	})
}

func TestClientHandlers_GetClient(t *testing.T) {
	handlers, testDB, userID := setupClientTest(t)
	defer testDB.Cleanup(t)

	// Create test client
	clientID := testDB.InsertTestClient(t, "Test Client")

	t.Run("get existing client", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%d", clientID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", clientID)},
		}

		handlers.GetClient(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		client, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Test Client", client["name"])
		assert.Equal(t, float64(clientID), client["id"])
	})

	t.Run("get nonexistent client", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.GetClient(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get client with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/invalid", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "invalid"},
		}

		handlers.GetClient(rr, req, ps)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestClientHandlers_CreateClient(t *testing.T) {
	handlers, testDB, userID := setupClientTest(t)
	defer testDB.Cleanup(t)

	t.Run("create valid client", func(t *testing.T) {
		phone := "555-1234"
		notes := "Test notes"
		reqBody := ClientRequest{
			Name:       "New Client",
			Email:      "newclient@example.com",
			Phone:      &phone,
			HourlyRate: 100.00,
			Notes:      &notes,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		client, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "New Client", client["name"])
		assert.Equal(t, "newclient@example.com", client["email"])
		assert.Equal(t, float64(100), client["hourlyRate"])
	})

	t.Run("create client without name", func(t *testing.T) {
		reqBody := ClientRequest{
			Name:       "",
			Email:      "test@example.com",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Equal(t, ErrCodeValidation, resp.Error.Code)
	})

	t.Run("create client with invalid email", func(t *testing.T) {
		reqBody := ClientRequest{
			Name:       "Test Client",
			Email:      "invalid-email",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create client with negative hourly rate", func(t *testing.T) {
		reqBody := ClientRequest{
			Name:       "Test Client",
			Email:      "test@example.com",
			HourlyRate: -50.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create client with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestClientHandlers_UpdateClient(t *testing.T) {
	handlers, testDB, userID := setupClientTest(t)
	defer testDB.Cleanup(t)

	// Create test client
	clientID := testDB.InsertTestClient(t, "Original Client")

	t.Run("update existing client", func(t *testing.T) {
		notes := "Updated notes"
		reqBody := ClientRequest{
			Name:       "Updated Client",
			Email:      "updated@example.com",
			HourlyRate: 150.00,
			Notes:      &notes,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/clients/%d", clientID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", clientID)},
		}

		handlers.UpdateClient(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		client, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Updated Client", client["name"])
		assert.Equal(t, "updated@example.com", client["email"])
		assert.Equal(t, float64(150), client["hourlyRate"])
	})

	t.Run("update nonexistent client", func(t *testing.T) {
		reqBody := ClientRequest{
			Name:       "Updated Client",
			Email:      "test@example.com",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/99999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.UpdateClient(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestClientHandlers_DeleteClient(t *testing.T) {
	handlers, testDB, userID := setupClientTest(t)
	defer testDB.Cleanup(t)

	// Create test client
	clientID := testDB.InsertTestClient(t, "Client to Delete")

	t.Run("delete existing client", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/clients/%d", clientID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", clientID)},
		}

		handlers.DeleteClient(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data["message"], "deleted successfully")
	})

	t.Run("delete nonexistent client", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/clients/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.DeleteClient(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestClientHandlers_GetClientHourlyRate(t *testing.T) {
	handlers, testDB, userID := setupClientTest(t)
	defer testDB.Cleanup(t)

	// Create test client with specific hourly rate
	clientID := testDB.InsertTestClientWithDefaults(t, "Rate Test Client", 125.00, "", "", "", "")

	t.Run("get client hourly rate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%d/hourlyrate", clientID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", clientID)},
		}

		handlers.GetClientHourlyRate(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(125), data["hourlyRate"])
	})
}
