package service

import (
	"github.com/pavel-trbv/go-todo-app/internal/domain"
	"github.com/pavel-trbv/go-todo-app/internal/repository"
)

type Authorization interface {
	CreateUser(user *domain.User) (int, error)
}

type TodoList interface {
}

type TodoItem interface {
}

type Service struct {
	Authorization
	TodoList
	TodoItem
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
	}
}
