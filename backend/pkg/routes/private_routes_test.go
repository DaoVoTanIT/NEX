package routes

import (
	"context"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/create-go-app/fiber-go-template/app/controllers"
	"github.com/create-go-app/fiber-go-template/app/dto"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/create-go-app/fiber-go-template/pkg/middleware"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

type authStub struct{}

func (s *authStub) SignUp(ctx context.Context, input *models.SignUp) (*core.ApiResponse, error) {
	return core.Success(200, "ok", nil, nil), nil
}
func (s *authStub) SignIn(ctx context.Context, input *models.SignIn) (*core.ApiResponse, error) {
	return core.Success(200, "ok", nil, nil), nil
}
func (s *authStub) SignOut(ctx context.Context, c any) (*core.ApiResponse, error) {
	return core.Success(204, "signed out", nil, nil), nil
}

type tokenStub struct{}

func (s *tokenStub) Renew(ctx context.Context, c any, refreshToken string) (*core.ApiResponse, error) {
	return core.Success(200, "ok", nil, nil), nil
}

type taskStub struct{}

func (s *taskStub) GetTasks(ctx context.Context) (*core.ApiResponse, error) {
	return core.Success(200, "ok", []dto.TaskRes{}, nil), nil
}
func (s *taskStub) GetTask(ctx context.Context, id string) (*core.ApiResponse, error) {
	return core.Success(200, "ok", dto.TaskRes{}, nil), nil
}
func (s *taskStub) Create(ctx context.Context, c any, req *dto.CreateTaskReq) (*core.ApiResponse, error) {
	return core.Success(200, "ok", nil, nil), nil
}
func (s *taskStub) Update(ctx context.Context, c any, task *models.Task) (*core.ApiResponse, error) {
	return core.Success(201, "updated", nil, nil), nil
}
func (s *taskStub) Delete(ctx context.Context, c any, id string) (*core.ApiResponse, error) {
	return core.Success(204, "deleted", nil, nil), nil
}

func TestPrivateRoutes(t *testing.T) {
	// Load .env.test file from the root folder.
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic(err)
	}

	// Create a sample data string.
	dataString := `{"id": "00000000-0000-0000-0000-000000000000"}`

	// Create token with `book:delete` credential.
	tokenOnlyDelete, err := utils.GenerateNewTokens(
		uuid.NewString(),
		[]string{"book:delete"},
	)
	if err != nil {
		panic(err)
	}

	// Create token without any credentials.
	tokenNoAccess, err := utils.GenerateNewTokens(
		uuid.NewString(),
		[]string{},
	)
	if err != nil {
		panic(err)
	}

	// Define a structure for specifying input and output data of a single test case.
	tests := []struct {
		description   string
		route         string // input route
		method        string // input method
		tokenString   string // input token
		body          io.Reader
		expectedError bool
		expectedCode  int
	}{
		{
			description:   "delete book without JWT and body (route missing)",
			route:         "/api/v1/book",
			method:        "DELETE",
			tokenString:   "",
			body:          nil,
			expectedError: false,
			expectedCode:  404,
		},
		{
			description:   "delete book without right credentials (route missing)",
			route:         "/api/v1/book",
			method:        "DELETE",
			tokenString:   "Bearer " + tokenNoAccess.Access,
			body:          strings.NewReader(dataString),
			expectedError: false,
			expectedCode:  404,
		},
		{
			description:   "delete book with credentials",
			route:         "/api/v1/book",
			method:        "DELETE",
			tokenString:   "Bearer " + tokenOnlyDelete.Access,
			body:          strings.NewReader(dataString),
			expectedError: false,
			expectedCode:  404,
		},
	}

	// Define a new Fiber app.
	app := fiber.New()

	// Define routes.
	authCtrl := controllers.NewAuthController(&authStub{})
	tokenCtrl := controllers.NewTokenController(&tokenStub{})
	taskCtrl := controllers.NewTaskController(&taskStub{})
	jwtMiddleware := middleware.NewJWTProtected(middleware.JWTConfig{
		SecretKey: os.Getenv("JWT_SECRET_KEY"),
	})
	PrivateRoutes(app, jwtMiddleware, authCtrl, tokenCtrl, taskCtrl)

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route from the test case.
		req := httptest.NewRequest(test.method, test.route, test.body)
		req.Header.Set("Authorization", test.tokenString)
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
