package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tinyauth/internal/api"
	"tinyauth/internal/auth"
	"tinyauth/internal/docker"
	"tinyauth/internal/hooks"
	"tinyauth/internal/providers"
	"tinyauth/internal/types"

	"github.com/magiconair/properties/assert"
)

// Simple API config for tests
var apiConfig = types.APIConfig{
	Port:            8080,
	Address:         "0.0.0.0",
	Secret:          "super-secret-api-thing-for-tests", // It is 32 chars long
	AppURL:          "http://tinyauth.localhost",
	CookieSecure:    false,
	SessionExpiry:   3600,
	DisableContinue: false,
}

// Cookie
var cookie string

// User
var user = types.User{
	Username: "user",
	Password: "$2a$10$AvGHLTYv3xiRJ0xV9xs3XeVIlkGTygI9nqIamFYB5Xu.5.0UWF7B6", // pass
}

// We need all this to be able to test the API
func getAPI(t *testing.T) *api.API {
	// Create docker service
	docker := docker.NewDocker()

	// Initialize docker
	dockerErr := docker.Init()

	// Check if there was an error
	if dockerErr != nil {
		t.Fatalf("Failed to initialize docker: %v", dockerErr)
	}

	// Create auth service
	auth := auth.NewAuth(docker, types.Users{
		{
			Username: user.Username,
			Password: user.Password,
		},
	}, nil, apiConfig.SessionExpiry)

	// Create providers service
	providers := providers.NewProviders(types.OAuthConfig{})

	// Initialize providers
	providers.Init()

	// Create hooks service
	hooks := hooks.NewHooks(auth, providers)

	// Create API
	api := api.NewAPI(apiConfig, hooks, auth, providers)

	// Setup routes
	api.Init()
	api.SetupRoutes()

	return api
}

// Test login (we will need this for the other tests)
func TestLogin(t *testing.T) {
	t.Log("Testing login")

	// Get API
	api := getAPI(t)

	// Create recorder
	recorder := httptest.NewRecorder()

	// Create request
	user := types.LoginRequest{
		Username: "user",
		Password: "pass",
	}

	json, err := json.Marshal(user)

	// Check if there was an error
	if err != nil {
		t.Fatalf("Error marshalling json: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "/api/login", strings.NewReader(string(json)))

	// Check if there was an error
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	// Serve the request
	api.Router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, recorder.Code, http.StatusOK)

	// Get the cookie
	cookie = recorder.Result().Cookies()[0].Value

	// Check if the cookie is set
	if cookie == "" {
		t.Fatalf("Cookie not set")
	}
}

// Test status
func TestStatus(t *testing.T) {
	t.Log("Testing status")

	// Get API
	api := getAPI(t)

	// Create recorder
	recorder := httptest.NewRecorder()

	// Create request
	req, err := http.NewRequest("GET", "/api/status", nil)

	// Check if there was an error
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	// Set the cookie
	req.AddCookie(&http.Cookie{
		Name:  "tinyauth",
		Value: cookie,
	})

	// Serve the request
	api.Router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, recorder.Code, http.StatusOK)

	// Parse the body
	body := recorder.Body.String()

	if !strings.Contains(body, "user") {
		t.Fatalf("Expected user in body")
	}
}

// Test logout
func TestLogout(t *testing.T) {
	t.Log("Testing logout")

	// Get API
	api := getAPI(t)

	// Create recorder
	recorder := httptest.NewRecorder()

	// Create request
	req, err := http.NewRequest("POST", "/api/logout", nil)

	// Check if there was an error
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	// Set the cookie
	req.AddCookie(&http.Cookie{
		Name:  "tinyauth",
		Value: cookie,
	})

	// Serve the request
	api.Router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, recorder.Code, http.StatusOK)

	// Check if the cookie is different (means go sessions flushed it)
	if recorder.Result().Cookies()[0].Value == cookie {
		t.Fatalf("Cookie not flushed")
	}
}

// TODO: Testing for the oauth stuff
