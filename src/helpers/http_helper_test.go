package helpers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pchchv/aas/src/constants"
	mocksData "github.com/pchchv/aas/src/database/mocks"
	"github.com/pchchv/aas/src/mocks"
	"github.com/pchchv/aas/src/models"
	"github.com/pchchv/aas/src/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRenderTemplateToBuffer(t *testing.T) {
	templateFS := &mocks.TestFS{
		FileContents: map[string]string{
			"layouts/layout.html": "<html>{{template \"content\" .}}</html>",
			"page.html":           "{{define \"content\"}}Hello, {{if .loggedInUser}}{{.loggedInUser.Username}}{{else}}Guest{{end}}!{{end}}",
		},
	}
	database := mocksData.NewDatabase(t)
	httpHelper := NewHttpHelper(templateFS, database)

	t.Run("Without ID Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, constants.ContextKeySettings, &models.Settings{AppName: "TestApp", UITheme: "light"})
		req = req.WithContext(ctx)
		data := map[string]interface{}{}
		buf, err := httpHelper.RenderTemplateToBuffer(req, "layouts/layout.html", "page.html", data)

		assert.NoError(t, err)
		assert.NotNil(t, buf)
		assert.Contains(t, buf.String(), "Hello, Guest!")
	})

	t.Run("With ID Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, constants.ContextKeySettings, &models.Settings{AppName: "TestApp", UITheme: "light"})
		// Mock JwtInfo with ID Token
		mockUser := &models.User{Id: 1, Username: "JohnDoe"}
		mockDatabase := mocksData.NewDatabase(t)
		mockDatabase.On("GetUserBySubject", mock.Anything, "user123").Return(mockUser, nil)
		jwtInfo := oauth.JwtInfo{
			IdToken: &oauth.Jwt{
				Claims: map[string]interface{}{
					"sub": "user123",
				},
			},
		}

		ctx = context.WithValue(ctx, constants.ContextKeyJwtInfo, jwtInfo)
		req = req.WithContext(ctx)
		httpHelper.database = mockDatabase
		data := map[string]interface{}{}
		buf, err := httpHelper.RenderTemplateToBuffer(req, "layouts/layout.html", "page.html", data)

		assert.NoError(t, err)
		assert.NotNil(t, buf)
		assert.Contains(t, buf.String(), "Hello, JohnDoe!")

		mockDatabase.AssertExpectations(t)
	})
}

func TestRenderTemplate(t *testing.T) {
	templateFS := &mocks.TestFS{
		FileContents: map[string]string{
			"layouts/layout.html": "<html>{{template \"content\" .}}</html>",
			"page.html":           "{{define \"content\"}}Hello, {{.Name}}! Status: {{._httpStatus}}{{end}}",
		},
	}
	database := mocksData.NewDatabase(t)
	httpHelper := NewHttpHelper(templateFS, database)

	t.Run("Without _httpStatus", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		ctx := req.Context()
		ctx = context.WithValue(ctx, constants.ContextKeySettings, &models.Settings{AppName: "TestApp", UITheme: "light"})
		req = req.WithContext(ctx)
		data := map[string]interface{}{
			"Name": "John",
		}

		err := httpHelper.RenderTemplate(w, req, "layouts/layout.html", "page.html", data)

		assert.NoError(t, err)
		assert.Equal(t, "text/html; charset=UTF-8", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "Hello, John!")
		assert.Contains(t, w.Body.String(), "Status:")
		assert.Equal(t, http.StatusOK, w.Code) // Default status should be 200 OK
	})

	t.Run("With _httpStatus", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		ctx := req.Context()
		ctx = context.WithValue(ctx, constants.ContextKeySettings, &models.Settings{AppName: "TestApp", UITheme: "light"})
		req = req.WithContext(ctx)
		data := map[string]interface{}{
			"Name":        "Jane",
			"_httpStatus": http.StatusCreated,
		}

		err := httpHelper.RenderTemplate(w, req, "layouts/layout.html", "page.html", data)

		assert.NoError(t, err)
		assert.Equal(t, "text/html; charset=UTF-8", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "Hello, Jane!")
		assert.Contains(t, w.Body.String(), "Status: 201")
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestGetFromUrlQueryOrFormPost(t *testing.T) {
	templateFS := &mocks.TestFS{}
	database := mocksData.NewDatabase(t)
	httpHelper := NewHttpHelper(templateFS, database)
	t.Run("Get from URL query", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?key=value", nil)
		value := httpHelper.GetFromUrlQueryOrFormPost(req, "key")
		assert.Equal(t, "value", value)
	})

	t.Run("Get from form post", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString("key=value"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		err := req.ParseForm()
		assert.NoError(t, err)
		value := httpHelper.GetFromUrlQueryOrFormPost(req, "key")
		assert.Equal(t, "value", value)
	})
}
