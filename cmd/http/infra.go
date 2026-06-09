package main

import (
	"context"
	"strconv"

	"github.com/AndreeJait/server-management-be/adapter/outbound"
	"github.com/AndreeJait/server-management-be/config"
	portOutbound "github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/server-management-be/pkg/retry"
	"github.com/AndreeJait/server-management-be/usecase"
	"github.com/AndreeJait/go-utility/v2/authw"
	"github.com/AndreeJait/go-utility/v2/jwtw"
	"github.com/AndreeJait/go-utility/v2/logw"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

// provideInfrastructure registers infrastructure providers into the dig container.
func provideInfrastructure(c *dig.Container) {
	c.Provide(newDB)
	c.Provide(newGormDB)
	c.Provide(newRedisConn)
	c.Provide(newRedisClient)
	c.Provide(newJWT)
	c.Provide(newAuthenticator)
	c.Provide(newRBAC)
	c.Provide(newDockerConn)
	c.Provide(newDockerEngine)
}

func newDB(cfg *config.AppConfig, cc *CleanupCollector) (*outbound.DB, error) {
	var db *outbound.DB
	var cleanup func(ctx context.Context) error
	err := retry.Do(context.Background(), retry.Config{
		MaxAttempts: cfg.Retry.MaxAttempts,
		Interval:    cfg.Retry.Interval,
		MaxInterval: cfg.Retry.MaxInterval,
	}, func() error {
		var connErr error
		db, cleanup, connErr = outbound.ConnectSQL(context.Background(), cfg)
		return connErr
	})
	if err != nil {
		return nil, err
	}
	cc.Add(cleanup)
	return db, nil
}

func newGormDB(db *outbound.DB) *gorm.DB {
	return db.GormDB
}

func newRedisConn(cfg *config.AppConfig, cc *CleanupCollector) (*outbound.RedisConn, error) {
	var conn *outbound.RedisConn
	var cleanup func(ctx context.Context) error
	err := retry.Do(context.Background(), retry.Config{
		MaxAttempts: cfg.Retry.MaxAttempts,
		Interval:    cfg.Retry.Interval,
		MaxInterval: cfg.Retry.MaxInterval,
	}, func() error {
		var connErr error
		conn, cleanup, connErr = outbound.ConnectRedis(context.Background(), cfg)
		return connErr
	})
	if err != nil {
		return nil, err
	}
	cc.Add(cleanup)
	return conn, nil
}

func newRedisClient(conn *outbound.RedisConn) *redis.Client {
	return conn.Client
}

func newJWT(cfg *config.AppConfig) jwtw.JWT {
	return jwtw.New(&jwtw.Config{SecretKey: cfg.Auth.JWTSecret})
}

func newAuthenticator(j jwtw.JWT) authw.Authenticator {
	return authw.NewJWT(&authw.JWTConfig{
		JWT:           j,
		NewClaims:     usecase.NewClaimsFactory(),
		ExtractResult: usecase.ExtractJWTClaims,
	})
}

func newRBAC(cfg *config.AppConfig, redisClient *redis.Client, userRepo portOutbound.UserRepository, roleRepo portOutbound.RoleRepository) *authw.RBAC {
	rbac := authw.NewRBAC(&authw.RBACConfig{
		RoleFetcher: func(ctx context.Context, userID string) ([]string, error) {
			id, err := strconv.ParseUint(userID, 10, 64)
			if err != nil {
				return nil, err
			}
			return userRepo.FindRolesByUserID(ctx, uint(id))
		},
		PermissionFetcher: func(ctx context.Context, userID string) ([]string, error) {
			roles, err := userRepo.FindRolesByUserID(ctx, func() uint {
				id, _ := strconv.ParseUint(userID, 10, 64)
				return uint(id)
			}())
			if err != nil {
				return nil, err
			}
			var allPerms []string
			seen := make(map[string]struct{})
			for _, r := range roles {
				perms, err := roleRepo.FindPermissionsByRole(ctx, r)
				if err != nil {
					return nil, err
				}
				for _, p := range perms {
					if _, ok := seen[p]; !ok {
						seen[p] = struct{}{}
						allPerms = append(allPerms, p)
					}
				}
			}
			return allPerms, nil
		},
		Cache:    authw.NewRedisCache(redisClient),
		CacheTTL: 600,
	})

	// Load role-permission mappings from DB; fall back to hardcoded defaults if table is empty
	registerRolesFromDB(rbac, roleRepo)
	return rbac
}

func registerRolesFromDB(rbac *authw.RBAC, roleRepo portOutbound.RoleRepository) {
	ctx := context.Background()
	rolePerms, err := roleRepo.FindAllPermissions(ctx)
	if err != nil {
		logw.Warningf("authw: failed to load role permissions from DB, using defaults: %v", err)
		registerDefaultRoles(rbac)
		return
	}

	if len(rolePerms) == 0 {
		registerDefaultRoles(rbac)
		return
	}

	for roleName, permissions := range rolePerms {
		rbac.RegisterRole(roleName, permissions...)
	}
	logw.Infof("Loaded %d roles from DB", len(rolePerms))
}

// registerDefaultRoles seeds the RBAC engine with hardcoded role-permission mappings.
// Used when the role_permissions table is empty (e.g. before migration runs).
func registerDefaultRoles(rbac *authw.RBAC) {
	rbac.RegisterRole("admin",
		"users:read", "users:write", "users:delete",
		"projects:read", "projects:write", "projects:delete",
		"apps:read", "apps:write", "apps:delete", "apps:deploy",
		"configs:read", "configs:write",
		"cloudflare:read", "cloudflare:write",
		"proxy:read", "proxy:write",
		"ssh:read", "ssh:write", "ssh:connect",
	)
	rbac.RegisterRole("operator",
		"users:read",
		"projects:read", "projects:write",
		"apps:read", "apps:write", "apps:deploy",
		"configs:read", "configs:write",
		"cloudflare:read", "cloudflare:write",
		"proxy:read",
		"ssh:read", "ssh:write", "ssh:connect",
	)
	rbac.RegisterRole("viewer",
		"users:read",
		"projects:read",
		"apps:read",
		"configs:read",
		"cloudflare:read",
		"proxy:read",
		"ssh:read",
	)
}

func newDockerConn(cfg *config.AppConfig, cc *CleanupCollector) (*outbound.DockerConn, error) {
	conn, err := outbound.ConnectDocker(cfg.Docker.Host)
	if err != nil {
		logw.Warningf("failed to connect Docker daemon: %v", err)
		return nil, err
	}
	cc.Add(outbound.DisconnectDocker(conn))
	return conn, nil
}

func newDockerEngine(conn *outbound.DockerConn) portOutbound.DockerEngine {
	return conn.Engine
}