package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/constants"
	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/repository"
)

type PipelineService struct {
	pipelineRepo *repository.PipelineRepository
}

func NewPipelineService(pipelineRepo *repository.PipelineRepository) *PipelineService {
	return &PipelineService{
		pipelineRepo: pipelineRepo,
	}
}

func (s *PipelineService) CreateBranch(ctx context.Context, projectID primitive.ObjectID, req *domain.CreateBranchRequest) (*domain.Branch, error) {
	var files []domain.CodeFile

	if req.CreatedFromRelease != nil {
		release, err := s.pipelineRepo.FindReleaseByVersion(ctx, projectID, *req.CreatedFromRelease)
		if err != nil {
			return nil, err
		}
		files = release.Files
	} else if req.ParentBranch != nil {
		parent, err := s.pipelineRepo.FindBranchByName(ctx, projectID, *req.ParentBranch)
		if err != nil {
			return nil, err
		}
		files = parent.Files
	} else {
		// Default: single main file with default code
		files = []domain.CodeFile{
			{Name: "main", Code: constants.DefaultServiceCode},
		}
	}

	branch := &domain.Branch{
		ProjectID:          projectID,
		Name:               req.Name,
		Files:              files,
		ParentBranch:       req.ParentBranch,
		CreatedFromRelease: req.CreatedFromRelease,
	}

	if err := s.pipelineRepo.CreateBranch(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

func (s *PipelineService) GetBranch(ctx context.Context, projectID primitive.ObjectID, name string) (*domain.Branch, error) {
	return s.pipelineRepo.FindBranchByName(ctx, projectID, name)
}

func (s *PipelineService) GetBranchByID(ctx context.Context, projectID primitive.ObjectID, branchID primitive.ObjectID) (*domain.Branch, error) {
	branch, err := s.pipelineRepo.FindBranchByID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if branch.ProjectID != projectID {
		return nil, repository.ErrBranchNotFound
	}
	return branch, nil
}

func (s *PipelineService) GetBranches(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Branch, error) {
	return s.pipelineRepo.FindBranchesByProject(ctx, projectID)
}

func (s *PipelineService) GetBranchSummaries(ctx context.Context, projectID primitive.ObjectID) ([]*domain.BranchSummary, error) {
	return s.pipelineRepo.FindBranchSummariesByProject(ctx, projectID)
}

func (s *PipelineService) UpdateBranch(ctx context.Context, projectID primitive.ObjectID, name string, req *domain.UpdateBranchRequest) (*domain.Branch, error) {
	branch, err := s.pipelineRepo.FindBranchByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	if err := s.validateFiles(req.Files); err != nil {
		return nil, err
	}

	branch.Files = req.Files

	if err := s.pipelineRepo.UpdateBranch(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

func (s *PipelineService) UpdateBranchByID(ctx context.Context, projectID primitive.ObjectID, branchID primitive.ObjectID, req *domain.UpdateBranchRequest) (*domain.Branch, error) {
	branch, err := s.GetBranchByID(ctx, projectID, branchID)
	if err != nil {
		return nil, err
	}

	if err := s.validateFiles(req.Files); err != nil {
		return nil, err
	}

	branch.Files = req.Files

	if err := s.pipelineRepo.UpdateBranch(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

// validateFiles checks that files array is valid
func (s *PipelineService) validateFiles(files []domain.CodeFile) error {
	if len(files) == 0 {
		return fmt.Errorf("at least one file is required")
	}

	hasMain := false
	names := make(map[string]bool)

	for _, f := range files {
		if f.Name == "" {
			return fmt.Errorf("file name cannot be empty")
		}
		if f.Name == "main" {
			hasMain = true
		}
		if names[f.Name] {
			return fmt.Errorf("duplicate file name: %s", f.Name)
		}
		names[f.Name] = true
	}

	if !hasMain {
		return fmt.Errorf("main file is required")
	}

	return nil
}

func (s *PipelineService) ResetBranch(ctx context.Context, projectID primitive.ObjectID, name string, req *domain.ResetBranchRequest) (*domain.Branch, error) {
	branch, err := s.pipelineRepo.FindBranchByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	release, err := s.pipelineRepo.FindReleaseByVersion(ctx, projectID, req.TargetVersion)
	if err != nil {
		return nil, err
	}

	branch.Files = release.Files
	branch.CreatedFromRelease = &req.TargetVersion

	if err := s.pipelineRepo.UpdateBranch(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

func (s *PipelineService) ResetBranchByID(ctx context.Context, projectID primitive.ObjectID, branchID primitive.ObjectID, req *domain.ResetBranchRequest) (*domain.Branch, error) {
	branch, err := s.GetBranchByID(ctx, projectID, branchID)
	if err != nil {
		return nil, err
	}

	release, err := s.pipelineRepo.FindReleaseByVersion(ctx, projectID, req.TargetVersion)
	if err != nil {
		return nil, err
	}

	branch.Files = release.Files
	branch.CreatedFromRelease = &req.TargetVersion

	if err := s.pipelineRepo.UpdateBranch(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

func (s *PipelineService) DeleteBranch(ctx context.Context, projectID primitive.ObjectID, name string) error {
	branch, err := s.pipelineRepo.FindBranchByName(ctx, projectID, name)
	if err != nil {
		return err
	}
	return s.pipelineRepo.DeleteBranch(ctx, branch.ID)
}

func (s *PipelineService) DeleteBranchByID(ctx context.Context, projectID primitive.ObjectID, branchID primitive.ObjectID) error {
	branch, err := s.GetBranchByID(ctx, projectID, branchID)
	if err != nil {
		return err
	}
	if branch.Name == "develop" {
		return fmt.Errorf("cannot delete develop branch")
	}
	return s.pipelineRepo.DeleteBranch(ctx, branchID)
}

func (s *PipelineService) CreateRelease(ctx context.Context, projectID primitive.ObjectID, req *domain.CreateReleaseRequest) (*domain.Release, error) {
	branch, err := s.pipelineRepo.FindBranchByName(ctx, projectID, req.BranchName)
	if err != nil {
		return nil, err
	}

	version, err := s.getNextVersion(ctx, projectID, req.BumpType)
	if err != nil {
		return nil, err
	}

	release := &domain.Release{
		ProjectID: projectID,
		Version:   version,
		Files:     branch.Files,
		Comment:   req.Comment,
		Tag:       req.Tag,
		IsActive:  false,
	}

	if err := s.pipelineRepo.CreateRelease(ctx, release); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *PipelineService) GetReleases(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Release, error) {
	return s.pipelineRepo.FindReleasesByProject(ctx, projectID)
}

func (s *PipelineService) GetReleaseSummaries(ctx context.Context, projectID primitive.ObjectID) ([]*domain.ReleaseSummary, error) {
	return s.pipelineRepo.FindReleaseSummariesByProject(ctx, projectID)
}

func (s *PipelineService) GetRelease(ctx context.Context, projectID primitive.ObjectID, version string) (*domain.Release, error) {
	return s.pipelineRepo.FindReleaseByVersion(ctx, projectID, version)
}

func (s *PipelineService) GetActiveRelease(ctx context.Context, projectID primitive.ObjectID) (*domain.Release, error) {
	return s.pipelineRepo.FindActiveRelease(ctx, projectID)
}

func (s *PipelineService) ActivateRelease(ctx context.Context, projectID primitive.ObjectID, version string) error {
	return s.pipelineRepo.ActivateRelease(ctx, projectID, version)
}

func (s *PipelineService) ActivateReleaseByID(ctx context.Context, projectID primitive.ObjectID, releaseID primitive.ObjectID) error {
	release, err := s.pipelineRepo.FindReleaseByID(ctx, releaseID)
	if err != nil {
		return err
	}
	if release.ProjectID != projectID {
		return repository.ErrReleaseNotFound
	}
	return s.pipelineRepo.ActivateRelease(ctx, projectID, release.Version)
}

func (s *PipelineService) DeleteRelease(ctx context.Context, projectID primitive.ObjectID, version string) error {
	release, err := s.pipelineRepo.FindReleaseByVersion(ctx, projectID, version)
	if err != nil {
		return err
	}

	if release.IsActive {
		return fmt.Errorf("cannot delete active release")
	}

	return s.pipelineRepo.DeleteRelease(ctx, release.ID)
}

func (s *PipelineService) DeleteReleaseByID(ctx context.Context, projectID primitive.ObjectID, releaseID primitive.ObjectID) error {
	release, err := s.pipelineRepo.FindReleaseByID(ctx, releaseID)
	if err != nil {
		return err
	}
	if release.ProjectID != projectID {
		return repository.ErrReleaseNotFound
	}
	if release.IsActive {
		return fmt.Errorf("cannot delete active release")
	}
	return s.pipelineRepo.DeleteRelease(ctx, releaseID)
}

func (s *PipelineService) getNextVersion(ctx context.Context, projectID primitive.ObjectID, bumpType string) (string, error) {
	latest, err := s.pipelineRepo.FindLatestRelease(ctx, projectID)
	if err != nil {
		if err == repository.ErrReleaseNotFound {
			return "1.0", nil
		}
		return "", err
	}

	parts := strings.Split(latest.Version, ".")
	if len(parts) != 2 {
		return "1.0", nil
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])

	if bumpType == "major" {
		major++
		minor = 0
	} else {
		minor++
	}

	return fmt.Sprintf("%d.%d", major, minor), nil
}

// File operations

// AddFile adds a new file to a branch
func (s *PipelineService) AddFile(ctx context.Context, projectID, branchID primitive.ObjectID, req *domain.CreateFileRequest) (*domain.Branch, error) {
	branch, err := s.GetBranchByID(ctx, projectID, branchID)
	if err != nil {
		return nil, err
	}

	// Check if file already exists
	if branch.GetFile(req.Name) != nil {
		return nil, fmt.Errorf("file already exists: %s", req.Name)
	}

	// Add file with empty code
	file := domain.CodeFile{Name: req.Name, Code: ""}
	if err := s.pipelineRepo.AddFileToBranch(ctx, branchID, file); err != nil {
		return nil, err
	}

	// Return updated branch
	return s.GetBranchByID(ctx, projectID, branchID)
}

// UpdateFile updates a single file's code
func (s *PipelineService) UpdateFile(ctx context.Context, projectID, branchID primitive.ObjectID, fileName string, req *domain.UpdateFileRequest) error {
	branch, err := s.GetBranchByID(ctx, projectID, branchID)
	if err != nil {
		return err
	}

	// Check if file exists
	if branch.GetFile(fileName) == nil {
		return fmt.Errorf("file not found: %s", fileName)
	}

	return s.pipelineRepo.UpdateBranchFile(ctx, branchID, fileName, req.Code)
}

// DeleteFile removes a file from a branch
func (s *PipelineService) DeleteFile(ctx context.Context, projectID, branchID primitive.ObjectID, fileName string) (*domain.Branch, error) {
	// Cannot delete main file
	if fileName == "main" {
		return nil, fmt.Errorf("cannot delete main file")
	}

	branch, err := s.GetBranchByID(ctx, projectID, branchID)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if branch.GetFile(fileName) == nil {
		return nil, fmt.Errorf("file not found: %s", fileName)
	}

	if err := s.pipelineRepo.DeleteFileFromBranch(ctx, branchID, fileName); err != nil {
		return nil, err
	}

	// Return updated branch
	return s.GetBranchByID(ctx, projectID, branchID)
}

// RenameFile renames a file in a branch
func (s *PipelineService) RenameFile(ctx context.Context, projectID, branchID primitive.ObjectID, fileName string, req *domain.RenameFileRequest) (*domain.Branch, error) {
	// Cannot rename main file
	if fileName == "main" {
		return nil, fmt.Errorf("cannot rename main file")
	}

	branch, err := s.GetBranchByID(ctx, projectID, branchID)
	if err != nil {
		return nil, err
	}

	// Check if source file exists
	if branch.GetFile(fileName) == nil {
		return nil, fmt.Errorf("file not found: %s", fileName)
	}

	// Check if target name is available
	if branch.GetFile(req.NewName) != nil {
		return nil, fmt.Errorf("file already exists: %s", req.NewName)
	}

	if err := s.pipelineRepo.RenameFileInBranch(ctx, branchID, fileName, req.NewName); err != nil {
		return nil, err
	}

	// Return updated branch
	return s.GetBranchByID(ctx, projectID, branchID)
}

// EnsureDevelopBranch creates develop branch if it doesn't exist
func (s *PipelineService) EnsureDevelopBranch(ctx context.Context, projectID primitive.ObjectID) (*domain.Branch, error) {
	branch, err := s.pipelineRepo.FindBranchByName(ctx, projectID, "develop")
	if err == nil {
		return branch, nil
	}

	if err != repository.ErrBranchNotFound {
		return nil, err
	}

	return s.CreateBranch(ctx, projectID, &domain.CreateBranchRequest{
		Name: "develop",
	})
}
