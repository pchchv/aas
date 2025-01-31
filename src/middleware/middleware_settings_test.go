package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pchchv/aas/src/constants"
	"github.com/pchchv/aas/src/database/mocks"
	"github.com/pchchv/aas/src/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMiddlewareSettings(t *testing.T) {
	t.Run("Successful retrieval of settings", func(t *testing.T) {
		mockDB := mocks.NewDatabase(t)
		expectedSettings := &models.Settings{
			Id:      1,
			AppName: "TestApp",
		}
		mockDB.On("GetSettingsById", mock.Anything, int64(1)).Return(expectedSettings, nil)

		middleware := MiddlewareSettings(mockDB)
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		var contextSettings *models.Settings
		middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextSettings = r.Context().Value(constants.ContextKeySettings).(*models.Settings)
		})).ServeHTTP(rr, req)

		assert.Equal(t, expectedSettings, contextSettings)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Database error", func(t *testing.T) {
		mockDB := mocks.NewDatabase(t)
		mockDB.On("GetSettingsById", mock.Anything, int64(1)).Return(nil, errors.New("database error"))

		middleware := MiddlewareSettings(mockDB)
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "fatal failure in GetSettings() middleware")
	})
}
