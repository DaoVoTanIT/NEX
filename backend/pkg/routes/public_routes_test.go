package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/create-go-app/fiber-go-template/app/controllers"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

type stubAuthService struct{}

func (s *stubAuthService) SignUp(ctx context.Context, input *models.SignUp) (*core.ApiResponse, error) {
	return core.Success(200, "ok", nil, nil), nil
}
func (s *stubAuthService) SignIn(ctx context.Context, input *models.SignIn) (*core.ApiResponse, error) {
	return core.Success(200, "ok", nil, nil), nil
}
func (s *stubAuthService) SignOut(ctx context.Context, c any) (*core.ApiResponse, error) {
	return core.Success(204, "signed out", nil, nil), nil
}

func TestPublicRoutes(t *testing.T) {
	// Load .env.test file from the root folder
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic(err)
	}

	// Define a structure for specifying input and output data of a single test case.
	tests := []struct {
		description   string
		route         string // input route
		expectedError bool
		expectedCode  int
	}{
		{
			description:   "get book by ID",
			route:         "/api/v1/book/" + uuid.New().String(),
			expectedError: false,
			expectedCode:  404,
		},
		{
			description:   "get book by invalid ID (non UUID)",
			route:         "/api/v1/book/123456",
			expectedError: false,
			expectedCode:  404,
		},
	}

	// Define Fiber app.
	app := fiber.New()

	// Create stub auth service and controller
	authCtrl := controllers.NewAuthController(&stubAuthService{})

	// Define routes.
	PublicRoutes(app, authCtrl)

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route from the test case.
		req := httptest.NewRequest("GET", test.route, http.NoBody)
		req.Header.Set("Content-Type", "application/json")

		// Perform the request plain with the app.
		resp, err := app.Test(req, -1) // the -1 disables request latency

		// Verify, that no error occurred, that is not expected
		assert.Equalf(t, test.expectedError, err != nil, test.description)

		// As expected errors lead to broken responses,
		// the next test case needs to be processed.
		if test.expectedError {
			continue
		}

		// Verify, if the status code is as expected.
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)
	}
}
