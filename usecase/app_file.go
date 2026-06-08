package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/appfile"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/logw"
)

type appFileUseCase struct {
	appFileRepo outbound.AppFileRepository
	appRepo     outbound.AppRepository
	deployRepo  outbound.DeploymentRepository
	filesystem  outbound.Filesystem
	deployFunc  appfile.DeployFunc
	hostBase    string
}

func NewAppFileUseCase(appFileRepo outbound.AppFileRepository, appRepo outbound.AppRepository, deployRepo outbound.DeploymentRepository, filesystem outbound.Filesystem, hostBase string) appfile.UseCase {
	return &appFileUseCase{appFileRepo: appFileRepo, appRepo: appRepo, deployRepo: deployRepo, filesystem: filesystem, hostBase: hostBase}
}

func (u *appFileUseCase) SetDeployFunc(fn appfile.DeployFunc) {
	u.deployFunc = fn
}

func (u *appFileUseCase) effectiveHostBase(app *entity.App) string {
	if app.BasePath != "" {
		return app.BasePath
	}
	return u.hostBase
}

func (u *appFileUseCase) validateOwnership(ctx context.Context, projectID uint, appID string) (*entity.App, error) {
	id, err := strconv.ParseUint(appID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid app ID")
	}
	a, err := u.appRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
	}
	if a.ProjectID != projectID {
		return nil, domainError.ErrForbidden.WithCustomMessage("App does not belong to this project")
	}
	return a, nil
}

func (u *appFileUseCase) triggerRedeploy(ctx context.Context, appID string) {
	if u.deployFunc == nil {
		return
	}
	app, err := u.appRepo.FindByAppID(ctx, appID)
	if err != nil {
		logw.Warningf("auto-redeploy: failed to find app %s: %v", appID, err)
		return
	}
	latest, err := u.deployRepo.FindLatestByAppID(ctx, appID)
	if err != nil {
		logw.Warningf("auto-redeploy: no previous deployment found for app %s: %v", appID, err)
		return
	}
	go func() {
		if _, err := u.deployFunc(context.Background(), appID, app.DeployToken, latest.Image); err != nil {
			logw.Warningf("auto-redeploy: failed for app %s: %v", appID, err)
		}
	}()
}

func (u *appFileUseCase) hostFilePath(app *entity.App, fileRelPath string) string {
	return filepath.Join(u.effectiveHostBase(app), app.AppID, "files", strings.TrimPrefix(fileRelPath, "/"))
}

func (u *appFileUseCase) writeToHost(app *entity.App, f *entity.AppFile) {
	hostPath := u.hostFilePath(app, f.Path)
	dir := filepath.Dir(hostPath)
	if err := u.filesystem.MkdirAll(dir, 0755); err != nil {
		logw.Warningf("failed to create directory %s: %v", dir, err)
		return
	}
	if f.FileType == "text" && f.Content != "" {
		if err := u.filesystem.WriteFile(hostPath, []byte(f.Content), 0644); err != nil {
			logw.Warningf("failed to write file %s: %v", hostPath, err)
		}
	}
}

func (u *appFileUseCase) Create(ctx context.Context, projectID uint, appID, path, content string) (*entity.AppFileResponse, error) {
	app, err := u.validateOwnership(ctx, projectID, appID)
	if err != nil {
		return nil, err
	}

	f := entity.NewAppFile(appID, path, content)
	if err := u.appFileRepo.Create(ctx, f); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	u.writeToHost(app, f)
	return f.ToResponse(), nil
}

func (u *appFileUseCase) List(ctx context.Context, projectID uint, appID string) ([]*entity.AppFileResponse, error) {
	if _, err := u.validateOwnership(ctx, projectID, appID); err != nil {
		return nil, err
	}

	files, err := u.appFileRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.AppFileResponse, 0, len(files))
	for _, f := range files {
		responses = append(responses, f.ToResponse())
	}
	return responses, nil
}

func (u *appFileUseCase) Get(ctx context.Context, projectID uint, appID string, fileID uint) (*entity.AppFileResponse, error) {
	if _, err := u.validateOwnership(ctx, projectID, appID); err != nil {
		return nil, err
	}

	f, err := u.appFileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("File not found")
	}
	if f.AppID != appID {
		return nil, domainError.ErrForbidden.WithCustomMessage("File does not belong to this app")
	}
	return f.ToResponse(), nil
}

func (u *appFileUseCase) Update(ctx context.Context, projectID uint, appID string, fileID uint, path, content string) (*entity.AppFileResponse, error) {
	app, err := u.validateOwnership(ctx, projectID, appID)
	if err != nil {
		return nil, err
	}

	f, err := u.appFileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("File not found")
	}
	if f.AppID != appID {
		return nil, domainError.ErrForbidden.WithCustomMessage("File does not belong to this app")
	}

	f.Path = path
	f.Content = content
	if err := u.appFileRepo.Update(ctx, f); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	u.writeToHost(app, f)
	return f.ToResponse(), nil
}

func (u *appFileUseCase) Delete(ctx context.Context, projectID uint, appID string, fileID uint) error {
	app, err := u.validateOwnership(ctx, projectID, appID)
	if err != nil {
		return err
	}

	f, err := u.appFileRepo.FindByID(ctx, fileID)
	if err != nil {
		return domainError.ErrNotFound.WithCustomMessage("File not found")
	}
	if f.AppID != appID {
		return domainError.ErrForbidden.WithCustomMessage("File does not belong to this app")
	}

	hostPath := u.hostFilePath(app, f.Path)
	_ = u.filesystem.RemoveFile(hostPath)

	if err := u.appFileRepo.Delete(ctx, fileID); err != nil {
		return domainError.ErrInternalServer.WithError(err)
	}
	return nil
}

func (u *appFileUseCase) Upload(ctx context.Context, projectID uint, appID, path string, data []byte, mimeType string, fileSize int64) (*entity.AppFileResponse, error) {
	app, err := u.validateOwnership(ctx, projectID, appID)
	if err != nil {
		return nil, err
	}

	cleanPath := strings.TrimPrefix(path, "/")
	if cleanPath == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("File path is required")
	}
	if strings.Contains(cleanPath, "..") {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Path traversal is not allowed")
	}

	hostPath := u.hostFilePath(app, cleanPath)
	dir := filepath.Dir(hostPath)
	if err := u.filesystem.MkdirAll(dir, 0755); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	if err := u.filesystem.WriteFile(hostPath, data, 0644); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	isText := strings.HasPrefix(mimeType, "text/") || mimeType == "application/json" || mimeType == "application/xml" || mimeType == "application/x-yaml"
	content := ""
	fileType := "binary"
	if isText {
		content = string(data)
		fileType = "text"
	}

	f := &entity.AppFile{
		AppID:    appID,
		Path:     cleanPath,
		Content:  content,
		FileType: fileType,
		FileSize: fileSize,
		MimeType: mimeType,
	}
	if err := u.appFileRepo.Create(ctx, f); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return f.ToResponse(), nil
}

func (u *appFileUseCase) Download(ctx context.Context, projectID uint, appID string, fileID uint) ([]byte, string, error) {
	app, err := u.validateOwnership(ctx, projectID, appID)
	if err != nil {
		return nil, "", err
	}

	f, err := u.appFileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, "", domainError.ErrNotFound.WithCustomMessage("File not found")
	}
	if f.AppID != appID {
		return nil, "", domainError.ErrForbidden.WithCustomMessage("File does not belong to this app")
	}

	hostPath := u.hostFilePath(app, f.Path)
	data, err := u.filesystem.ReadFile(hostPath)
	if err != nil {
		return nil, "", domainError.ErrInternalServer.WithError(err)
	}

	mimeType := f.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return data, mimeType, nil
}

func (u *appFileUseCase) CreateFolder(ctx context.Context, projectID uint, appID, path string) (*entity.AppResponse, error) {
	app, err := u.validateOwnership(ctx, projectID, appID)
	if err != nil {
		return nil, err
	}

	cleanPath := strings.Trim(strings.TrimPrefix(path, "/"), " ")
	if cleanPath == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Folder path is required")
	}
	if strings.Contains(cleanPath, "..") {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Path traversal is not allowed")
	}

	hostPath := filepath.Join(u.effectiveHostBase(app), app.AppID, cleanPath)
	if err := u.filesystem.MkdirAll(hostPath, 0755); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	containerPath := "/" + cleanPath
	volumeMount := entity.VolumeMount{
		HostPath:      hostPath,
		ContainerPath: containerPath,
		Mode:          "rw",
	}

	app.VolumeMounts = append(app.VolumeMounts, volumeMount)
	if err := u.appRepo.Update(ctx, app); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	return app.ToResponse(), nil
}

func (u *appFileUseCase) DeleteFolder(ctx context.Context, projectID uint, appID, path string) error {
	app, err := u.validateOwnership(ctx, projectID, appID)
	if err != nil {
		return err
	}

	cleanPath := strings.Trim(strings.TrimPrefix(path, "/"), " ")
	if cleanPath == "" {
		return domainError.ErrInvalidParam.WithCustomMessage("Folder path is required")
	}
	if strings.Contains(cleanPath, "..") {
		return domainError.ErrInvalidParam.WithCustomMessage("Path traversal is not allowed")
	}

	hostPath := filepath.Join(u.effectiveHostBase(app), app.AppID, cleanPath)
	containerPath := "/" + cleanPath

	if err := u.filesystem.RemoveAll(hostPath); err != nil {
		logw.Warningf("failed to remove folder %s: %v", hostPath, err)
	}

	mountKey := fmt.Sprintf("%s:%s", hostPath, containerPath)
	filtered := make(entity.VolumeMountList, 0, len(app.VolumeMounts))
	for _, vm := range app.VolumeMounts {
		vmKey := fmt.Sprintf("%s:%s", vm.HostPath, vm.ContainerPath)
		if vmKey != mountKey {
			filtered = append(filtered, vm)
		}
	}
	app.VolumeMounts = filtered

	if err := u.appRepo.Update(ctx, app); err != nil {
		return domainError.ErrInternalServer.WithError(err)
	}

	return nil
}