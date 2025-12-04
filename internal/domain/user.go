package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password" json:"-"`
	Name        string             `bson:"name" json:"name"`
	Avatar      string             `bson:"avatar" json:"avatar"`
	IsRoot      bool               `bson:"is_root" json:"is_root"`
	Permissions Permissions        `bson:"permissions" json:"permissions"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type Permissions struct {
	CreateProjects bool                 `bson:"create_projects" json:"create_projects"`
	ManageUsers    bool                 `bson:"manage_users" json:"manage_users"`
	ProjectAccess  []primitive.ObjectID `bson:"project_access" json:"project_access"`
}

type CreateUserRequest struct {
	Email       string      `json:"email" binding:"required,email"`
	Password    string      `json:"password" binding:"required,min=6"`
	Name        string      `json:"name" binding:"required"`
	Permissions Permissions `json:"permissions"`
}

type UpdateUserRequest struct {
	Email       *string      `json:"email" binding:"omitempty,email"`
	Name        *string      `json:"name"`
	Permissions *Permissions `json:"permissions"`
}

type UpdateProfileRequest struct {
	Name   *string `json:"name"`
	Avatar *string `json:"avatar"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
