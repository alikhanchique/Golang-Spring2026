package usecase

import (
	"alikhan.practice3/internal/repository"
	"alikhan.practice3/pkg/modules"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) GetUsers() ([]modules.User, error) {
	return u.repo.GetUsers()
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
	return u.repo.GetUserByID(id)
}

func (u *UserUsecase) CreateUser(req modules.CreateUserRequest) (int, error) {
	return u.repo.CreateUser(req)
}

func (u *UserUsecase) UpdateUser(id int, req modules.UpdateUserRequest) error {
	return u.repo.UpdateUser(id, req)
}

func (u *UserUsecase) DeleteUser(id int) (int64, error) {
	return u.repo.DeleteUser(id)
}

func (u *UserUsecase) Healthcheck() string {
	return "ok"
}
