package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReleaseTag string

const (
	ReleaseTagStable     ReleaseTag = "stable"
	ReleaseTagHotFix     ReleaseTag = "hot-fix"
	ReleaseTagNightBuild ReleaseTag = "night-build"
	ReleaseTagDevelop    ReleaseTag = "develop"
)

type Branch struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID          primitive.ObjectID `bson:"project_id" json:"project_id"`
	Name               string             `bson:"name" json:"name"`
	Code               string             `bson:"code" json:"code"`
	ParentBranch       *string            `bson:"parent_branch" json:"parent_branch"`
	CreatedFromRelease *string            `bson:"created_from_release" json:"created_from_release"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

type Release struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	Version   string             `bson:"version" json:"version"`
	Code      string             `bson:"code" json:"code"`
	Comment   string             `bson:"comment" json:"comment"`
	Tag       ReleaseTag         `bson:"tag" json:"tag"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type CreateBranchRequest struct {
	Name               string  `json:"name" binding:"required"`
	ParentBranch       *string `json:"parent_branch"`
	CreatedFromRelease *string `json:"created_from_release"`
}

type UpdateBranchRequest struct {
	Code string `json:"code" binding:"required"`
}

type CreateReleaseRequest struct {
	BranchName string     `json:"branch_name" binding:"required"`
	BumpType   string     `json:"bump_type" binding:"required,oneof=major minor"`
	Comment    string     `json:"comment"`
	Tag        ReleaseTag `json:"tag" binding:"required"`
}

type ResetBranchRequest struct {
	TargetVersion string `json:"target_version" binding:"required"`
}
