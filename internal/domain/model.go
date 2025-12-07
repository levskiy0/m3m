package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FieldType string

const (
	FieldTypeString   FieldType = "string"
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeFloat    FieldType = "float"
	FieldTypeBool     FieldType = "bool"
	FieldTypeDocument FieldType = "document"
	FieldTypeFile     FieldType = "file"
	FieldTypeRef      FieldType = "ref"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeSelect   FieldType = "select"
)

type Model struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	Name        string             `bson:"name" json:"name"`
	Slug        string             `bson:"slug" json:"slug"`
	Fields      []ModelField       `bson:"fields" json:"fields"`
	TableConfig TableConfig        `bson:"table_config" json:"table_config"`
	FormConfig  FormConfig         `bson:"form_config" json:"form_config"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type ModelField struct {
	Key          string      `bson:"key" json:"key"`
	Type         FieldType   `bson:"type" json:"type"`
	Required     bool        `bson:"required" json:"required"`
	DefaultValue interface{} `bson:"default_value" json:"default_value"`
	RefModel     string      `bson:"ref_model,omitempty" json:"ref_model,omitempty"`
	Options      []string    `bson:"options,omitempty" json:"options,omitempty"`
}

type TableConfig struct {
	Columns     []string `bson:"columns" json:"columns"`
	Filters     []string `bson:"filters" json:"filters"`
	SortColumns []string `bson:"sort_columns" json:"sort_columns"`
	Searchable  []string `bson:"searchable" json:"searchable"`
}

type FormConfig struct {
	FieldOrder   []string          `bson:"field_order" json:"field_order"`
	HiddenFields []string          `bson:"hidden_fields" json:"hidden_fields"`
	FieldViews   map[string]string `bson:"field_views" json:"field_views"`
}

type CreateModelRequest struct {
	Name        string       `json:"name" binding:"required"`
	Slug        string       `json:"slug" binding:"required"`
	Fields      []ModelField `json:"fields" binding:"required"`
	TableConfig *TableConfig `json:"table_config"`
	FormConfig  *FormConfig  `json:"form_config"`
}

type UpdateModelRequest struct {
	Name        *string       `json:"name"`
	Fields      *[]ModelField `json:"fields"`
	TableConfig *TableConfig  `json:"table_config"`
	FormConfig  *FormConfig   `json:"form_config"`
}

type ModelData struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ModelID   primitive.ObjectID     `bson:"_model_id" json:"_model_id"`
	Data      map[string]interface{} `bson:",inline" json:"data"`
	CreatedAt time.Time              `bson:"_created_at" json:"created_at"`
	UpdatedAt time.Time              `bson:"_updated_at" json:"updated_at"`
}

type DataQuery struct {
	Page    int               `form:"page"`
	Limit   int               `form:"limit"`
	Sort    string            `form:"sort"`
	Order   string            `form:"order"`
	Filters map[string]string `form:"-"` // Parsed from query params
}

// FilterOperator represents a filter operation
type FilterOperator string

const (
	FilterOpEquals     FilterOperator = "eq"
	FilterOpNotEquals  FilterOperator = "ne"
	FilterOpGreater    FilterOperator = "gt"
	FilterOpGreaterEq  FilterOperator = "gte"
	FilterOpLess       FilterOperator = "lt"
	FilterOpLessEq     FilterOperator = "lte"
	FilterOpContains   FilterOperator = "contains"
	FilterOpStartsWith FilterOperator = "startsWith"
	FilterOpEndsWith   FilterOperator = "endsWith"
	FilterOpIn         FilterOperator = "in"
	FilterOpNotIn      FilterOperator = "notIn"
	FilterOpIsNull     FilterOperator = "isNull"
	FilterOpIsNotNull  FilterOperator = "isNotNull"
)

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
}

// AdvancedDataQuery supports complex filtering
type AdvancedDataQuery struct {
	Page     int               `json:"page" form:"page"`
	Limit    int               `json:"limit" form:"limit"`
	Sort     string            `json:"sort" form:"sort"`
	Order    string            `json:"order" form:"order"`
	Filters  []FilterCondition `json:"filters"`
	Search   string            `json:"search" form:"search"`     // Full-text search query
	SearchIn []string          `json:"searchIn" form:"searchIn"` // Fields to search in
}
