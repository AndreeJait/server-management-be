package outbound

import (
	"context"
	"time"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type domainRequestCountRepository struct {
	db *DB
}

func NewDomainRequestCountRepository(db *DB) *domainRequestCountRepository {
	return &domainRequestCountRepository{db: db}
}

func (r *domainRequestCountRepository) Increment(ctx context.Context, domain string) error {
	now := time.Now().UTC()
	return r.db.GormDB.WithContext(ctx).Exec(
		`INSERT INTO domain_request_counts (domain, count, last_request_at, created_at, updated_at)
		 VALUES (?, 1, ?, ?, ?)
		 ON CONFLICT (domain) DO UPDATE SET count = domain_request_counts.count + 1, last_request_at = ?, updated_at = ?`,
		domain, now, now, now, now, now,
	).Error
}

func (r *domainRequestCountRepository) FindAll(ctx context.Context) ([]*entity.DomainRequestCount, error) {
	var counts []*entity.DomainRequestCount
	if err := r.db.GormDB.WithContext(ctx).Order("count DESC").Find(&counts).Error; err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *domainRequestCountRepository) Reset(ctx context.Context, domain string) error {
	return r.db.GormDB.WithContext(ctx).Exec(
		`UPDATE domain_request_counts SET count = 0, updated_at = ? WHERE domain = ?`,
		time.Now().UTC(), domain,
	).Error
}