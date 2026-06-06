package usecase

import (
	"context"
	"strconv"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/registry"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type registryUseCase struct {
	credRepo outbound.RegistryCredentialRepository
}

func NewRegistryUseCase(credRepo outbound.RegistryCredentialRepository) registry.UseCase {
	return &registryUseCase{credRepo: credRepo}
}

func (u *registryUseCase) Create(ctx context.Context, scope string, projectID *uint, registryURL, username, password string) (*entity.RegistryCredentialResponse, error) {
	var cred *entity.RegistryCredential
	switch entity.CredentialScope(scope) {
	case entity.CredentialScopeGlobal:
		cred = entity.NewGlobalRegistryCredential(registryURL, username, password)
	case entity.CredentialScopeProject:
		if projectID == nil {
			return nil, domainError.ErrInvalidParam.WithCustomMessage("project_id required for project-scoped credentials")
		}
		cred = entity.NewProjectRegistryCredential(*projectID, registryURL, username, password)
	default:
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid scope, must be 'global' or 'project'")
	}
	if err := u.credRepo.Create(ctx, cred); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return cred.ToResponse(), nil
}

func (u *registryUseCase) ListGlobal(ctx context.Context) ([]*entity.RegistryCredentialResponse, error) {
	creds, err := u.credRepo.FindGlobal(ctx)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.RegistryCredentialResponse, 0, len(creds))
	for _, c := range creds {
		responses = append(responses, c.ToResponse())
	}
	return responses, nil
}

func (u *registryUseCase) ListByProject(ctx context.Context, projectID uint) ([]*entity.RegistryCredentialResponse, error) {
	creds, err := u.credRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.RegistryCredentialResponse, 0, len(creds))
	for _, c := range creds {
		responses = append(responses, c.ToResponse())
	}
	return responses, nil
}

func (u *registryUseCase) Update(ctx context.Context, credID string, registryURL, username, password string) (*entity.RegistryCredentialResponse, error) {
	id, err := strconv.ParseUint(credID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid credential ID")
	}
	cred, err := u.credRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("Registry credential not found")
	}
	cred.RegistryURL = registryURL
	cred.Username = username
	cred.Password = password // TODO: encrypt at rest (Phase 3)
	if err := u.credRepo.Update(ctx, cred); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return cred.ToResponse(), nil
}

func (u *registryUseCase) Delete(ctx context.Context, credID string) error {
	id, err := strconv.ParseUint(credID, 10, 64)
	if err != nil {
		return domainError.ErrInvalidParam.WithCustomMessage("Invalid credential ID")
	}
	if err := u.credRepo.Delete(ctx, uint(id)); err != nil {
		return domainError.ErrInternalServer.WithError(err)
	}
	return nil
}