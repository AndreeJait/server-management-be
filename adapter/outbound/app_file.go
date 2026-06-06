package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type appFileRepository struct {
	db *gorm.DB
}

func NewAppFileRepository(db *DB) outbound.AppFileRepository {
	return &appFileRepository{db: db.GormDB}
}

func (r *appFileRepository) Create(ctx context.Context, file *entity.AppFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *appFileRepository) FindByID(ctx context.Context, id uint) (*entity.AppFile, error) {
	var file entity.AppFile
	if err := r.db.WithContext(ctx).First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *appFileRepository) FindByAppID(ctx context.Context, appID string) ([]*entity.AppFile, error) {
	var files []*entity.AppFile
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (r *appFileRepository) Update(ctx context.Context, file *entity.AppFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *appFileRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.AppFile{}, id).Error
}

func (r *appFileRepository) DeleteByAppID(ctx context.Context, appID string) error {
	return r.db.WithContext(ctx).Where("app_id = ?", appID).Delete(&entity.AppFile{}).Error
}