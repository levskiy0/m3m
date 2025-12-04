package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EnvVarType string

const (
	EnvVarTypeString  EnvVarType = "string"
	EnvVarTypeText    EnvVarType = "text"
	EnvVarTypeJSON    EnvVarType = "json"
	EnvVarTypeInteger EnvVarType = "integer"
	EnvVarTypeFloat   EnvVarType = "float"
	EnvVarTypeBoolean EnvVarType = "boolean"
)

type EnvVar struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	Key       string             `bson:"key" json:"key"`
	Type      EnvVarType         `bson:"type" json:"type"`
	Value     interface{}        `bson:"value" json:"value"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateEnvVarRequest struct {
	Key   string      `json:"key" binding:"required"`
	Type  EnvVarType  `json:"type" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

type UpdateEnvVarRequest struct {
	Type  *EnvVarType  `json:"type"`
	Value *interface{} `json:"value"`
}
