package service

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/repository"
)

var (
	ErrCannotDeleteRoot = errors.New("cannot delete root user")
	ErrCannotModifyRoot = errors.New("cannot modify root user permissions")
	ErrCannotBlockRoot  = errors.New("cannot block root user")
)

type UserService struct {
	userRepo    *repository.UserRepository
	authService *AuthService
}

func NewUserService(userRepo *repository.UserRepository, authService *AuthService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (s *UserService) Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	hashedPassword, err := s.authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:       req.Email,
		Password:    hashedPassword,
		Name:        req.Name,
		Permissions: req.Permissions,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) CreateRootUser(ctx context.Context, email, password string) (*domain.User, error) {
	// Check if root user already exists
	_, err := s.userRepo.FindRootUser(ctx)
	if err == nil {
		return nil, errors.New("root user already exists")
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	hashedPassword, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    email,
		Password: hashedPassword,
		Name:     "Root Admin",
		IsRoot:   true,
		Permissions: domain.Permissions{
			CreateProjects: true,
			ManageUsers:    true,
			ProjectAccess:  []primitive.ObjectID{},
		},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

func (s *UserService) GetAll(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *UserService) Update(ctx context.Context, id primitive.ObjectID, req *domain.UpdateUserRequest, currentUser *domain.User) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only root can modify root user
	if user.IsRoot && !currentUser.IsRoot {
		return nil, ErrCannotModifyRoot
	}

	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Permissions != nil && !user.IsRoot {
		user.Permissions = *req.Permissions
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id primitive.ObjectID, req *domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, id primitive.ObjectID, req *domain.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.authService.CheckPassword(user.Password, req.OldPassword) {
		return ErrInvalidCredentials
	}

	hashedPassword, err := s.authService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id primitive.ObjectID) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if user.IsRoot {
		return ErrCannotDeleteRoot
	}

	return s.userRepo.Delete(ctx, id)
}

func (s *UserService) Block(ctx context.Context, id primitive.ObjectID) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if user.IsRoot {
		return ErrCannotBlockRoot
	}

	user.IsBlocked = true
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) Unblock(ctx context.Context, id primitive.ObjectID) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	user.IsBlocked = false
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) IsBlocked(ctx context.Context, id primitive.ObjectID) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	return user.IsBlocked, nil
}
