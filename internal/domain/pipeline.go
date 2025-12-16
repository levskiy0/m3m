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

// CodeFile represents a single file in a branch or release
type CodeFile struct {
	Name string `bson:"name" json:"name"`
	Code string `bson:"code" json:"code"`
}

type Branch struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID          primitive.ObjectID `bson:"project_id" json:"project_id"`
	Name               string             `bson:"name" json:"name"`
	Files              []CodeFile         `bson:"files" json:"files"`
	ParentBranch       *string            `bson:"parent_branch" json:"parent_branch"`
	CreatedFromRelease *string            `bson:"created_from_release" json:"created_from_release"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

// GetFile returns a file by name, or nil if not found
func (b *Branch) GetFile(name string) *CodeFile {
	for i := range b.Files {
		if b.Files[i].Name == name {
			return &b.Files[i]
		}
	}
	return nil
}

// GetMainFile returns the main file (entry point)
func (b *Branch) GetMainFile() *CodeFile {
	return b.GetFile("main")
}

// BranchSummary is a lightweight version of Branch without Code for list operations
type BranchSummary struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID          primitive.ObjectID `bson:"project_id" json:"project_id"`
	Name               string             `bson:"name" json:"name"`
	ParentBranch       *string            `bson:"parent_branch" json:"parent_branch"`
	CreatedFromRelease *string            `bson:"created_from_release" json:"created_from_release"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

type Release struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	Version   string             `bson:"version" json:"version"`
	Files     []CodeFile         `bson:"files" json:"files"`
	Comment   string             `bson:"comment" json:"comment"`
	Tag       ReleaseTag         `bson:"tag" json:"tag"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// GetFile returns a file by name, or nil if not found
func (r *Release) GetFile(name string) *CodeFile {
	for i := range r.Files {
		if r.Files[i].Name == name {
			return &r.Files[i]
		}
	}
	return nil
}

// GetMainFile returns the main file (entry point)
func (r *Release) GetMainFile() *CodeFile {
	return r.GetFile("main")
}

// ReleaseSummary is a lightweight version of Release without Code for list operations
type ReleaseSummary struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	Version   string             `bson:"version" json:"version"`
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
	Files []CodeFile `json:"files" binding:"required,dive"`
}

// CreateFileRequest is used to create a new file in a branch
type CreateFileRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateFileRequest is used to update a single file's code
type UpdateFileRequest struct {
	Code string `json:"code" binding:"required"`
}

// RenameFileRequest is used to rename a file
type RenameFileRequest struct {
	NewName string `json:"new_name" binding:"required"`
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
