package ssh

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type UseCase interface {
	Create(ctx context.Context, name, host string, port int, username, authMethod, password, privateKey string, ownerID uint) (*entity.SSHHostResponse, error)
	List(ctx context.Context, ownerID uint) ([]*entity.SSHHostResponse, error)
	Get(ctx context.Context, id uint, ownerID uint) (*entity.SSHHostResponse, error)
	Update(ctx context.Context, id uint, ownerID uint, name, host string, port int, username, authMethod, password, privateKey string) (*entity.SSHHostResponse, error)
	Delete(ctx context.Context, id uint, ownerID uint) error
	Connect(ctx context.Context, hostID uint, ownerID uint) (outbound.SSHSession, error)
}