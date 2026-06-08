package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type DomainRequestCountRepository interface {
	Increment(ctx context.Context, domain string) error
	FindAll(ctx context.Context) ([]*entity.DomainRequestCount, error)
	Reset(ctx context.Context, domain string) error
}