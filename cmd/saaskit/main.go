// Package main is the entry point for the SaaSKit server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	_ "github.com/medaminerjb/saas-kit/docs" // swagger docs
	"github.com/medaminerjb/saas-kit/internal/audit"
	"github.com/medaminerjb/saas-kit/internal/config"
	idcrypto "github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/handler"
	"github.com/medaminerjb/saas-kit/internal/identity/repository"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
	oidcprovider "github.com/medaminerjb/saas-kit/internal/oidc/provider"
	"github.com/medaminerjb/saas-kit/internal/oidc/relyingparty"
	platformcrypto "github.com/medaminerjb/saas-kit/internal/platform/crypto"
	"github.com/medaminerjb/saas-kit/internal/platform/database"
	"github.com/medaminerjb/saas-kit/internal/platform/events"
	"github.com/medaminerjb/saas-kit/internal/platform/jobs"
	tenantrepo "github.com/medaminerjb/saas-kit/internal/tenant/repository"
	tenantservice "github.com/medaminerjb/saas-kit/internal/tenant/service"
)

// @title           SaaSKit API
// @version         1.0
// @description     Enterprise-grade Identity and Access Management platform with OIDC, multi-tenant support, and API key management.
// @termsOfService  https://github.com/medaminerjb/saas-kit/blob/main/TERMS.md

// @contact.name   SaaSKit Support
// @contact.url    https://github.com/medaminerjb/saas-kit/issues
// @contact.email  support@saaskit.io

// @license.name  MIT
// @license.url   https://github.com/medaminerjb/saas-kit/blob/main/LICENSE

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ─── Configuration ────────────────────────────────
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// ─── Logger ───────────────────────────────────────
	logger := setupLogger(cfg.Log)
	logger.Info("starting SaaSKit",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.Port),
	)

	// ─── Database ─────────────────────────────────────
	pool, err := database.Connect(ctx, database.Config{
		URL:      cfg.Database.DSN(),
		MaxConns: cfg.Database.MaxConns,
		MinConns: cfg.Database.MinConns,
	})
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()
	logger.Info("database connected")

	// ─── Run Migrations ───────────────────────────────
	if cfg.IsDevelopment() {
		logger.Info("running migrations (dev mode)")
		if err := runMigrations(cfg.Database.DSN()); err != nil {
			return fmt.Errorf("running migrations: %w", err)
		}
	}

	// ─── Signing Keys ─────────────────────────────────
	keyPair, err := idcrypto.LoadOrGenerateKeyPair(
		cfg.JWT.KeyPath,
		cfg.JWT.Algorithm,
		cfg.IsDevelopment(),
	)
	if err != nil {
		return fmt.Errorf("loading signing keys: %w", err)
	}
	logger.Info("signing keys loaded", slog.String("algorithm", keyPair.Algorithm))

	// ─── Crypto ───────────────────────────────────────
	hasher := idcrypto.NewHasher(idcrypto.Argon2Params{
		Memory:      cfg.Argon2.Memory,
		Iterations:  cfg.Argon2.Iterations,
		Parallelism: cfg.Argon2.Parallelism,
		SaltLength:  cfg.Argon2.SaltLength,
		KeyLength:   cfg.Argon2.KeyLength,
	})

	tokenHasher := idcrypto.NewTokenHasher(cfg.ServerSecret)

	// ─── Envelope Encryption ─────────────────────────
	var envelope *platformcrypto.Envelope
	if cfg.EncryptionMasterKey != "" {
		envelope, err = platformcrypto.NewEnvelope(cfg.EncryptionMasterKey)
		if err != nil {
			return fmt.Errorf("initializing envelope encryption: %w", err)
		}
		logger.Info("envelope encryption initialized")
	}

	// ─── Event Publisher ──────────────────────────────
	logPublisher := events.NewLogPublisher(logger)
	auditPublisher := audit.NewAuditPublisher(logger, pool)
	publisher := events.NewMultiPublisher(logPublisher, auditPublisher)

	// ─── Repositories ─────────────────────────────────
	userRepo := repository.NewUserRepo(pool)
	sessionRepo := repository.NewSessionRepo(pool)
	tokenRepo := repository.NewTokenRepo(pool)

	// ─── Super Admin Initialization ──────────────────
	if err := ensureSuperAdminFromConfig(ctx, userRepo, hasher, logger); err != nil {
		logger.Error("failed to ensure super admin user", slog.Any("error", err))
	}

	// ─── Services ─────────────────────────────────────
	tokenService := service.NewTokenService(service.TokenServiceConfig{
		PrivateKey: keyPair.PrivateKey,
		PublicKey:  keyPair.PublicKey,
		Algorithm:  keyPair.Algorithm,
		KeyID:      keyPair.KeyID,
		Issuer:     cfg.JWT.Issuer,
		AccessTTL:  cfg.JWT.AccessTTL,
	}, logger)

	authService := service.NewAuthService(service.AuthServiceConfig{
		Users:        userRepo,
		Sessions:     sessionRepo,
		Tokens:       tokenRepo,
		Hasher:       hasher,
		TokenHasher:  tokenHasher,
		TokenService: tokenService,
		Publisher:    publisher,
		RefreshTTL:   cfg.JWT.RefreshTTL,
	}, logger)

	userService := service.NewUserService(userRepo, publisher, logger)

	identityManager := service.NewIdentityManager(authService, userService, tokenService, logger)

	// ─── Tenant Service ────────────────────────────────
	tenantRepo := tenantrepo.NewTenantRepository(pool)
	tenantService := tenantservice.NewTenantService(tenantRepo, publisher, logger)

	// ─── OIDC Provider ────────────────────────────────
	oidcStorage := oidcprovider.NewStorage(oidcprovider.StorageConfig{
		Pool:     pool,
		UserRepo: userRepo,
		Hasher:   hasher,
		Envelope: envelope,
		KeyPair:  keyPair,
		Logger:   logger,
	})

	oidcHandler, err := oidcprovider.SetupProvider(cfg.BaseURL, oidcStorage, logger)
	if err != nil {
		return fmt.Errorf("setting up OIDC provider: %w", err)
	}

	// ─── Social Login ─────────────────────────────────
	socialRouter := chi.NewRouter()
	socialHandler := relyingparty.NewHandler(relyingparty.Config{
		GoogleClientID:     cfg.OAuth.Google.ClientID,
		GoogleClientSecret: cfg.OAuth.Google.ClientSecret,
		GitHubClientID:     cfg.OAuth.GitHub.ClientID,
		GitHubClientSecret: cfg.OAuth.GitHub.ClientSecret,
		BaseURL:            cfg.BaseURL,
		UserRepo:           userRepo,
		IdentityManager:    identityManager,
		Publisher:          publisher,
		Logger:             logger,
	})
	socialHandler.Routes(socialRouter)

	// ─── API Key Service ──────────────────────────────
	apiKeyRepo := repository.NewAPIKeyRepo(pool)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, publisher, logger)

	// ─── HTTP Server ──────────────────────────────────
	router := handler.NewRouter(handler.RouterConfig{
		Identity:      identityManager,
		Pool:          pool,
		Logger:        logger,
		OIDCProvider:  oidcHandler,
		SocialLogin:   socialRouter,
		TenantService: tenantService,
		APIKeyService: apiKeyService,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Background Jobs ──────────────────────────────
	scheduler := jobs.NewScheduler(logger)
	scheduler.Register(&cleanupJob{
		name:     "cleanup_sessions",
		interval: 1 * time.Hour,
		fn:       func(ctx context.Context) error { _, err := sessionRepo.DeleteExpired(ctx); return err },
	})
	scheduler.Register(&cleanupJob{
		name:     "cleanup_tokens",
		interval: 1 * time.Hour,
		fn:       func(ctx context.Context) error { _, err := tokenRepo.DeleteExpired(ctx); return err },
	})
	scheduler.Start(ctx)

	// ─── Start Server ─────────────────────────────────
	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// ─── Graceful Shutdown ────────────────────────────
	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		scheduler.Wait()
		logger.Info("server stopped")
	}

	return nil
}

// cleanupJob is a simple background job adapter.
type cleanupJob struct {
	name     string
	interval time.Duration
	fn       func(ctx context.Context) error
}

func (j *cleanupJob) Name() string                  { return j.name }
func (j *cleanupJob) Interval() time.Duration       { return j.interval }
func (j *cleanupJob) Run(ctx context.Context) error { return j.fn(ctx) }

func setupLogger(cfg config.LogConfig) *slog.Logger {
	var handler slog.Handler
	level := slog.LevelInfo
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func runMigrations(databaseURL string) error {
	// In development, we use goose programmatically.
	// For production, use `make migrate-up` or the goose CLI.
	// This is a placeholder — will wire goose.Up() when dependency is added.
	_ = databaseURL
	return nil
}

func ensureSuperAdminFromConfig(ctx context.Context, userRepo repository.UserRepository, hasher *idcrypto.Hasher, logger *slog.Logger) error {
	email := os.Getenv("SAASKIT_SUPERADMIN_EMAIL")
	if email == "" {
		return nil
	}

	if err := service.ValidateEmail(email); err != nil {
		return fmt.Errorf("invalid SAASKIT_SUPERADMIN_EMAIL: %w", err)
	}

	var password string
	passwordFile := os.Getenv("SAASKIT_SUPERADMIN_PASSWORD_FILE")
	if passwordFile != "" {
		data, err := os.ReadFile(passwordFile)
		if err != nil {
			return fmt.Errorf("reading SAASKIT_SUPERADMIN_PASSWORD_FILE: %w", err)
		}
		password = strings.TrimSpace(string(data))
	} else {
		password = os.Getenv("SAASKIT_SUPERADMIN_PASSWORD")
	}

	if password == "" {
		logger.Warn("SAASKIT_SUPERADMIN_EMAIL is set but no SAASKIT_SUPERADMIN_PASSWORD or SAASKIT_SUPERADMIN_PASSWORD_FILE provided")
		return nil
	}

	if err := service.ValidatePassword(password); err != nil {
		return fmt.Errorf("invalid SAASKIT_SUPERADMIN_PASSWORD: %w", err)
	}

	passwordHash, err := hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("hashing super admin password: %w", err)
	}

	existingUser, _ := userRepo.GetByEmail(ctx, email, nil)
	if existingUser != nil {
		existingUser.PasswordHash = &passwordHash
		existingUser.Status = domain.UserStatusActive
		existingUser.EmailVerified = true
		if existingUser.MetadataPrivate == nil {
			existingUser.MetadataPrivate = make(map[string]interface{})
		}
		existingUser.MetadataPrivate["role"] = "super_admin"
		existingUser.MetadataPrivate["is_super_admin"] = true

		if err := userRepo.Update(ctx, existingUser); err != nil {
			return fmt.Errorf("updating super admin user: %w", err)
		}
		logger.Info("super admin user updated successfully", slog.String("email", email))
		return nil
	}

	metaPrivate := map[string]interface{}{
		"role":           "super_admin",
		"is_super_admin": true,
	}

	newUser := &domain.User{
		ID:              uuid.New(),
		Email:           email,
		Name:            "Super Admin",
		PasswordHash:    &passwordHash,
		Status:          domain.UserStatusActive,
		EmailVerified:   true,
		MetadataPrivate: metaPrivate,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := userRepo.Create(ctx, newUser); err != nil {
		return fmt.Errorf("creating super admin user: %w", err)
	}

	logger.Info("super admin user created successfully", slog.String("email", email))
	return nil
}
