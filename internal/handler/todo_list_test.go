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

func withUserId(userId interface{}) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		if userId != nil {
			ctx.Set(userCtx, userId)
		}
	}
	return f
}

func TestHandler_createList(t *testing.T) {
	type mockBehavior func(s *mock_service.MockTodoList, userId interface{}, list domain.TodoList)

	testTable := []struct {
		name                string
		inputBody           string
		inputList           domain.TodoList
		userId              interface{}
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"title":"list title","description":"list desc"}`,
			inputList: domain.TodoList{
				Title:       "list title",
				Description: "list desc",
			},
			userId: 1,
			mockBehavior: func(s *mock_service.MockTodoList, userId interface{}, list domain.TodoList) {
				s.EXPECT().Create(userId, list).Return(1, nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: `{"id":1}`,
		},
		{
			name:                "Missing user id",
			mockBehavior:        func(s *mock_service.MockTodoList, userId interface{}, list domain.TodoList) {},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"user id not found"}`,
		},
		{
			name:                "Invalid type of user id",
			userId:              "1",
			mockBehavior:        func(s *mock_service.MockTodoList, userId interface{}, list domain.TodoList) {},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"user id is of invalid type"}`,
		},
		{
			name:                "Missing Fields",
			inputBody:           `{}`,
			userId:              1,
			mockBehavior:        func(s *mock_service.MockTodoList, userId interface{}, list domain.TodoList) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:                "Invalid input",
			inputBody:           `{"title":"list`,
			userId:              1,
			mockBehavior:        func(s *mock_service.MockTodoList, userId interface{}, list domain.TodoList) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service Error",
			inputBody: `{"title":"list title","description":"list desc"}`,
			inputList: domain.TodoList{
				Title:       "list title",
				Description: "list desc",
			},
			userId: 1,
			mockBehavior: func(s *mock_service.MockTodoList, userId interface{}, list domain.TodoList) {
				s.EXPECT().Create(userId, list).Return(0, errors.New("service error"))
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

			s := mock_service.NewMockTodoList(c)
			test.mockBehavior(s, test.userId, test.inputList)

			services := &service.Service{TodoList: s}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			r.Use(func(ctx *gin.Context) {
				if test.userId != nil {
					ctx.Set(userCtx, test.userId)
				}
			})
			r.POST("/api/lists", handler.createList)

			// Test Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/lists",
				bytes.NewBufferString(test.inputBody))

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.JSONEq(t, test.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_getAllLists(t *testing.T) {
	type mockBehavior func(s *mock_service.MockTodoList, userId interface{})

	lists := []domain.TodoList{
		{
			Id:          1,
			Title:       "list",
			Description: "desc",
		},
	}

	testTable := []struct {
		name                string
		userId              interface{}
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:   "OK",
			userId: 1,
			mockBehavior: func(s *mock_service.MockTodoList, userId interface{}) {
				s.EXPECT().GetAll(userId).Return(lists, nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: `{"data": [{ "id": 1, "title": "list", "description": "desc" }]}`,
		},
		{
			name:                "Missing user id",
			mockBehavior:        func(s *mock_service.MockTodoList, userId interface{}) {},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"user id not found"}`,
		},
		{
			name:                "Invalid type of user id",
			userId:              "1",
			mockBehavior:        func(s *mock_service.MockTodoList, userId interface{}) {},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"user id is of invalid type"}`,
		},
		{
			name:   "Service Error",
			userId: 1,
			mockBehavior: func(s *mock_service.MockTodoList, userId interface{}) {
				s.EXPECT().GetAll(userId).Return(nil, errors.New("service error"))
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

			s := mock_service.NewMockTodoList(c)
			test.mockBehavior(s, test.userId)

			services := &service.Service{TodoList: s}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			r.Use(withUserId(test.userId))
			r.GET("/api/lists", handler.getAllLists)

			// Test Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/lists", nil)

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.JSONEq(t, test.expectedRequestBody, w.Body.String())
		})
	}
}
