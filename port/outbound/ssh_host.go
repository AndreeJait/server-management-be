package outbound

import (
	"context"
	"io"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

// SSHSession represents an active SSH shell session.
type SSHSession interface {
	Stdin() io.Writer
	Stdout() io.Reader
	Close() error
	Resize(rows, cols int) error
}

// SSHHostRepository manages SSH host records in the database.
type SSHHostRepository interface {
	Create(ctx context.Context, host *entity.SSHHost) error
	FindByID(ctx context.Context, id uint) (*entity.SSHHost, error)
	FindByOwnerID(ctx context.Context, ownerID uint) ([]*entity.SSHHost, error)
	Update(ctx context.Context, host *entity.SSHHost) error
	Delete(ctx context.Context, id uint) error
}

// SSHClient opens interactive SSH sessions to remote hosts.
type SSHClient interface {
	Connect(ctx context.Context, host string, port int, username, authMethod, password, privateKey string) (SSHSession, error)
}