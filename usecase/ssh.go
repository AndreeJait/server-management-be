package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/ssh"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type sshUseCase struct {
	hostRepo  outbound.SSHHostRepository
	sshClient outbound.SSHClient
}

func NewSSHUseCase(hostRepo outbound.SSHHostRepository, sshClient outbound.SSHClient) ssh.UseCase {
	return &sshUseCase{hostRepo: hostRepo, sshClient: sshClient}
}

func (u *sshUseCase) Create(ctx context.Context, name, host string, port int, username, authMethod, password, privateKey string, ownerID uint) (*entity.SSHHostResponse, error) {
	am := entity.AuthMethod(authMethod)
	if !am.IsValid() {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid auth method. Use 'password' or 'private_key'")
	}
	if port <= 0 || port > 65535 {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Port must be between 1 and 65535")
	}
	h := entity.NewSSHHost(name, host, port, username, am, password, privateKey, ownerID)
	if err := u.hostRepo.Create(ctx, h); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return h.ToResponse(), nil
}

func (u *sshUseCase) List(ctx context.Context, ownerID uint) ([]*entity.SSHHostResponse, error) {
	hosts, err := u.hostRepo.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.SSHHostResponse, 0, len(hosts))
	for _, h := range hosts {
		responses = append(responses, h.ToResponse())
	}
	return responses, nil
}

func (u *sshUseCase) Get(ctx context.Context, id uint, ownerID uint) (*entity.SSHHostResponse, error) {
	h, err := u.hostRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("SSH host not found")
	}
	if h.OwnerID != ownerID {
		return nil, domainError.ErrForbidden.WithCustomMessage("SSH host does not belong to you")
	}
	return h.ToResponse(), nil
}

func (u *sshUseCase) Update(ctx context.Context, id uint, ownerID uint, name, host string, port int, username, authMethod, password, privateKey string) (*entity.SSHHostResponse, error) {
	h, err := u.hostRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("SSH host not found")
	}
	if h.OwnerID != ownerID {
		return nil, domainError.ErrForbidden.WithCustomMessage("SSH host does not belong to you")
	}
	am := entity.AuthMethod(authMethod)
	if !am.IsValid() {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid auth method. Use 'password' or 'private_key'")
	}
	if port <= 0 || port > 65535 {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Port must be between 1 and 65535")
	}
	h.Name = name
	h.Host = host
	h.Port = port
	h.Username = username
	h.AuthMethod = am
	h.Password = password
	h.PrivateKey = privateKey
	if err := u.hostRepo.Update(ctx, h); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return h.ToResponse(), nil
}

func (u *sshUseCase) Delete(ctx context.Context, id uint, ownerID uint) error {
	h, err := u.hostRepo.FindByID(ctx, id)
	if err != nil {
		return domainError.ErrNotFound.WithCustomMessage("SSH host not found")
	}
	if h.OwnerID != ownerID {
		return domainError.ErrForbidden.WithCustomMessage("SSH host does not belong to you")
	}
	return u.hostRepo.Delete(ctx, id)
}

func (u *sshUseCase) Connect(ctx context.Context, hostID uint, ownerID uint) (outbound.SSHSession, error) {
	h, err := u.hostRepo.FindByID(ctx, hostID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("SSH host not found")
	}
	if h.OwnerID != ownerID {
		return nil, domainError.ErrForbidden.WithCustomMessage("SSH host does not belong to you")
	}
	session, err := u.sshClient.Connect(ctx, h.Host, h.Port, h.Username, string(h.AuthMethod), h.Password, h.PrivateKey)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithCustomMessage(fmt.Sprintf("SSH connection failed: %s", err.Error()))
	}
	return session, nil
}

// userIDFromCtx extracts the numeric user ID from context set by auth middleware.
func userIDFromCtx(ctx context.Context) (uint, error) {
	uidStr, ok := ctx.Value("user_id").(string)
	if !ok || uidStr == "" {
		return 0, fmt.Errorf("user ID not found in context")
	}
	uid, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}
	return uint(uid), nil
}