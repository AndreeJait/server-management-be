package usecase

import (
	"context"
	"strconv"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type projectUseCase struct {
	projectRepo outbound.ProjectRepository
}

func NewProjectUseCase(projectRepo outbound.ProjectRepository) project.UseCase {
	return &projectUseCase{projectRepo: projectRepo}
}

func (u *projectUseCase) Create(ctx context.Context, name, description string, ownerID uint) (*entity.ProjectResponse, error) {
	p := entity.NewProject(name, description, ownerID)
	if err := u.projectRepo.Create(ctx, p); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return p.ToResponse(), nil
}

func (u *projectUseCase) List(ctx context.Context, ownerID uint) ([]*entity.ProjectResponse, error) {
	projects, err := u.projectRepo.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.ProjectResponse, 0, len(projects))
	for _, p := range projects {
		responses = append(responses, p.ToResponse())
	}
	return responses, nil
}

func (u *projectUseCase) Get(ctx context.Context, projectID string, ownerID uint) (*entity.ProjectResponse, error) {
	id, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid project ID")
	}
	p, err := u.projectRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("Project not found")
	}
	if p.OwnerID != ownerID {
		return nil, domainError.ErrForbidden.WithCustomMessage("Not your project")
	}
	return p.ToResponse(), nil
}

func (u *projectUseCase) Update(ctx context.Context, projectID string, ownerID uint, name, description string) (*entity.ProjectResponse, error) {
	id, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid project ID")
	}
	p, err := u.projectRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("Project not found")
	}
	if p.OwnerID != ownerID {
		return nil, domainError.ErrForbidden.WithCustomMessage("Not your project")
	}
	p.Name = name
	p.Description = description
	if err := u.projectRepo.Update(ctx, p); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return p.ToResponse(), nil
}

func (u *projectUseCase) Delete(ctx context.Context, projectID string, ownerID uint) error {
	id, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		return domainError.ErrInvalidParam.WithCustomMessage("Invalid project ID")
	}
	p, err := u.projectRepo.FindByID(ctx, uint(id))
	if err != nil {
		return domainError.ErrNotFound.WithCustomMessage("Project not found")
	}
	if p.OwnerID != ownerID {
		return domainError.ErrForbidden.WithCustomMessage("Not your project")
	}
	if err := u.projectRepo.Delete(ctx, uint(id)); err != nil {
		return domainError.ErrInternalServer.WithError(err)
	}
	return nil
}