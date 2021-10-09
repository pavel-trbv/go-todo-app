package handler

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/pavel-trbv/go-todo-app/internal/domain"
	"github.com/pavel-trbv/go-todo-app/internal/service"
	mock_service "github.com/pavel-trbv/go-todo-app/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestHandler_signUp(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAuthorization, user domain.User)

	testTable := []struct {
		name                string
		inputBody           string
		inputUser           domain.User
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"name":"Test","username":"test","password":"qwerty"}`,
			inputUser: domain.User{
				Name:     "Test",
				Username: "test",
				Password: "qwerty",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user domain.User) {
				s.EXPECT().CreateUser(user).Return(1, nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: `{"id":1}`,
		},
		{
			name:                "Missing Fields",
			inputBody:           `{"username":"test","password":"qwerty"}`,
			mockBehavior:        func(s *mock_service.MockAuthorization, user domain.User) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:                "Invalid Body",
			inputBody:           `{"username":`,
			mockBehavior:        func(s *mock_service.MockAuthorization, user domain.User) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service Failure",
			inputBody: `{"name":"Test","username":"test","password":"qwerty"}`,
			inputUser: domain.User{
				Name:     "Test",
				Username: "test",
				Password: "qwerty",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user domain.User) {
				s.EXPECT().CreateUser(user).Return(0, errors.New("service error"))
			},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"service error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehavior(auth, testCase.inputUser)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			r.POST("/sign-up", handler.signUp)

			// Test Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-up",
				bytes.NewBufferString(testCase.inputBody))

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_signIn(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAuthorization, user signInInput)

	testTable := []struct {
		name                string
		inputBody           string
		inputUser           signInInput
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"username":"test","password":"qwerty"}`,
			inputUser: signInInput{
				Username: "test",
				Password: "qwerty",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user signInInput) {
				s.EXPECT().GenerateToken(user.Username, user.Password).Return("token", nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: `{"token":"token"}`,
		},
		{
			name:                "Missing Fields",
			inputBody:           `{"username":"test"}`,
			mockBehavior:        func(s *mock_service.MockAuthorization, user signInInput) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:                "Invalid Body",
			inputBody:           `{"username":`,
			mockBehavior:        func(s *mock_service.MockAuthorization, user signInInput) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service Failure",
			inputBody: `{"username":"test","password":"qwerty"}`,
			inputUser: signInInput{
				Username: "test",
				Password: "qwerty",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user signInInput) {
				s.EXPECT().GenerateToken(user.Username, user.Password).
					Return("", errors.New("service error"))
			},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"service error"}`,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			test.mockBehavior(auth, test.inputUser)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			r.POST("/sign-in", handler.signIn)

			// Test Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-in",
				bytes.NewBufferString(test.inputBody))

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedRequestBody, w.Body.String())
		})
	}
}
