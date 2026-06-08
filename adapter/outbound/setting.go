package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type settingRepository struct {
	db *DB
}

func NewSettingRepository(db *DB) *settingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) FindAll(ctx context.Context) ([]*entity.Setting, error) {
	var settings []*entity.Setting
	if err := r.db.GormDB.WithContext(ctx).Order("section, key").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *settingRepository) FindBySection(ctx context.Context, section string) ([]*entity.Setting, error) {
	var settings []*entity.Setting
	if err := r.db.GormDB.WithContext(ctx).Where("section = ?", section).Order("key").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *settingRepository) FindByKey(ctx context.Context, section, key string) (*entity.Setting, error) {
	var s entity.Setting
	if err := r.db.GormDB.WithContext(ctx).Where("section = ? AND key = ?", section, key).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settingRepository) Upsert(ctx context.Context, setting *entity.Setting) error {
	return r.db.GormDB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "section"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(setting).Error
}

func (r *settingRepository) UpsertBatch(ctx context.Context, settings []*entity.Setting) error {
	return r.db.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, s := range settings {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "section"}, {Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
			}).Create(s).Error; err != nil {
				return err
			}
		}
		return nil
	})
}