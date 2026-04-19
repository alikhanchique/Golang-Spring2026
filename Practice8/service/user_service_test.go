package service

import (
	"errors"
	"practice-8/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)

	result, err := userService.GetUserByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := userService.CreateUser(user)
	assert.NoError(t, err)
}

func TestRegisterUser(t *testing.T) {
	t.Run("user already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		existing := &repository.User{ID: 2, Name: "Alice", Email: "alice@example.com"}
		// GetByEmail returns a non-nil user → email is taken
		mockRepo.EXPECT().GetByEmail("alice@example.com").Return(existing, nil)

		err := svc.RegisterUser(&repository.User{Name: "Bob"}, "alice@example.com")
		assert.EqualError(t, err, "user with this email already exists")
	})

	t.Run("new user success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		newUser := &repository.User{Name: "Bob", Email: "bob@example.com"}
		// GetByEmail returns nil user (email free) and nil error
		mockRepo.EXPECT().GetByEmail("bob@example.com").Return(nil, nil)
		mockRepo.EXPECT().CreateUser(newUser).Return(nil)

		err := svc.RegisterUser(newUser, "bob@example.com")
		assert.NoError(t, err)
	})

	t.Run("repository error on CreateUser", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		newUser := &repository.User{Name: "Carol", Email: "carol@example.com"}
		repoErr := errors.New("db connection failed")

		mockRepo.EXPECT().GetByEmail("carol@example.com").Return(nil, nil)
		mockRepo.EXPECT().CreateUser(newUser).Return(repoErr)

		err := svc.RegisterUser(newUser, "carol@example.com")
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestUpdateUserName(t *testing.T) {
	t.Run("empty name returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		err := svc.UpdateUserName(2, "")
		assert.EqualError(t, err, "name cannot be empty")
	})

	t.Run("user not found / repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		repoErr := errors.New("user not found")
		mockRepo.EXPECT().GetUserByID(99).Return(nil, repoErr)

		err := svc.UpdateUserName(99, "NewName")
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("successful update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		user := &repository.User{ID: 2, Name: "OldName"}

		mockRepo.EXPECT().GetUserByID(2).Return(user, nil)
		// Verify the name was actually changed before UpdateUser is called
		mockRepo.EXPECT().UpdateUser(gomock.AssignableToTypeOf(user)).
			DoAndReturn(func(u *repository.User) error {
				assert.Equal(t, "NewName", u.Name, "name should be updated before persistence")
				return nil
			})

		err := svc.UpdateUserName(2, "NewName")
		assert.NoError(t, err)
	})

	t.Run("UpdateUser fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		user := &repository.User{ID: 2, Name: "OldName"}
		repoErr := errors.New("db write error")

		mockRepo.EXPECT().GetUserByID(2).Return(user, nil)
		mockRepo.EXPECT().UpdateUser(gomock.Any()).Return(repoErr)

		err := svc.UpdateUserName(2, "NewName")
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("attempt to delete admin user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		// No repo call should happen when deleting admin
		err := svc.DeleteUser(1)
		assert.EqualError(t, err, "it is not allowed to delete admin user")
	})

	t.Run("successful delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		// Verify repo.DeleteUser is actually called with the correct id
		mockRepo.EXPECT().DeleteUser(5).Return(nil)

		err := svc.DeleteUser(5)
		assert.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		repoErr := errors.New("delete failed")
		mockRepo.EXPECT().DeleteUser(3).Return(repoErr)

		err := svc.DeleteUser(3)
		assert.ErrorIs(t, err, repoErr)
	})
}
