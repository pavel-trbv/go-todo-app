package service

import (
	"crypto/sha1"
	"fmt"
	"github.com/pavel-trbv/go-todo-app/internal/domain"
	"github.com/pavel-trbv/go-todo-app/internal/repository"
)

const salt = "fsdf7ashagbsv789sa11"

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(user *domain.User) (int, error) {
	user.Password = generatePasswordHash(user.Password)
	return s.repo.CreateUser(user)
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
