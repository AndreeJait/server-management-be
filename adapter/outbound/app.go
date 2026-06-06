package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type appRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *DB) outbound.AppRepository {
	return &appRepository{db: db.GormDB}
}

func (r *appRepository) Create(ctx context.Context, app *entity.App) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *appRepository) FindByID(ctx context.Context, id uint) (*entity.App, error) {
	var app entity.App
	if err := r.db.WithContext(ctx).First(&app, id).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) FindByProjectID(ctx context.Context, projectID uint) ([]*entity.App, error) {
	var apps []*entity.App
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *appRepository) FindByAppID(ctx context.Context, appID string) (*entity.App, error) {
	var app entity.App
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) Update(ctx context.Context, app *entity.App) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *appRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.App{}, id).Error
}