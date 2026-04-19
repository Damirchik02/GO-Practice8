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
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Bakytzhan"}
	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)

	result, err := svc.GetUserByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 2, Name: "Alice"}
	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := svc.CreateUser(user)
	assert.NoError(t, err)
}

func TestRegisterUser_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	existing := &repository.User{ID: 1, Email: "a@b.com"}
	mockRepo.EXPECT().GetByEmail("a@b.com").Return(existing, nil)

	err := svc.RegisterUser(&repository.User{Name: "Bob"}, "a@b.com")
	assert.ErrorContains(t, err, "already exists")
}

func TestRegisterUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{Name: "Bob", Email: "new@b.com"}
	mockRepo.EXPECT().GetByEmail("new@b.com").Return(nil, nil)
	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := svc.RegisterUser(user, "new@b.com")
	assert.NoError(t, err)
}

func TestRegisterUser_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	mockRepo.EXPECT().GetByEmail("x@y.com").Return(nil, errors.New("db error"))

	err := svc.RegisterUser(&repository.User{}, "x@y.com")
	assert.ErrorContains(t, err, "error getting user")
}

func TestUpdateUserName_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := NewUserService(repository.NewMockUserRepository(ctrl))

	err := svc.UpdateUserName(1, "")
	assert.ErrorContains(t, err, "name cannot be empty")
}

func TestUpdateUserName_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	mockRepo.EXPECT().GetUserByID(99).Return(nil, errors.New("not found"))

	err := svc.UpdateUserName(99, "NewName")
	assert.Error(t, err)
}

func TestUpdateUserName_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 2, Name: "Old"}
	mockRepo.EXPECT().GetUserByID(2).Return(user, nil)
	mockRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(u *repository.User) error {
		assert.Equal(t, "NewName", u.Name) // verify name was actually changed
		return nil
	})

	err := svc.UpdateUserName(2, "NewName")
	assert.NoError(t, err)
}

func TestUpdateUserName_UpdateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 3, Name: "Old"}
	mockRepo.EXPECT().GetUserByID(3).Return(user, nil)
	mockRepo.EXPECT().UpdateUser(gomock.Any()).Return(errors.New("update failed"))

	err := svc.UpdateUserName(3, "NewName")
	assert.Error(t, err)
}

func TestDeleteUser_Admin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := NewUserService(repository.NewMockUserRepository(ctrl))

	err := svc.DeleteUser(1)
	assert.ErrorContains(t, err, "admin")
}

func TestDeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	mockRepo.EXPECT().DeleteUser(5).DoAndReturn(func(id int) error {
		assert.Equal(t, 5, id) // verify correct user deleted
		return nil
	})

	err := svc.DeleteUser(5)
	assert.NoError(t, err)
}

func TestDeleteUser_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	mockRepo.EXPECT().DeleteUser(3).Return(errors.New("db error"))

	err := svc.DeleteUser(3)
	assert.Error(t, err)
}
