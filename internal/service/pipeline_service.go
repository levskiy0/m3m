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
	var code string

	if req.CreatedFromRelease != nil {
		release, err := s.pipelineRepo.FindReleaseByVersion(ctx, projectID, *req.CreatedFromRelease)
		if err != nil {
			return nil, err
		}
		code = release.Code
	} else if req.ParentBranch != nil {
		parent, err := s.pipelineRepo.FindBranchByName(ctx, projectID, *req.ParentBranch)
		if err != nil {
			return nil, err
		}
		code = parent.Code
	} else {
		code = constants.DefaultServiceCode
	}

	branch := &domain.Branch{
		ProjectID:          projectID,
		Name:               req.Name,
		Code:               code,
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

	branch.Code = req.Code

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

	branch.Code = req.Code

	if err := s.pipelineRepo.UpdateBranch(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
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

	branch.Code = release.Code
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

	branch.Code = release.Code
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
		Code:      branch.Code,
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
