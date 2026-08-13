package provider

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// SetupProvider creates and configures the OpenID Connect Provider.
func SetupProvider(issuer string, storage *Storage, logger *slog.Logger) (http.Handler, error) {
	config := &op.Config{
		CodeMethodS256:        true,
		AuthMethodPost:        true,
		AuthMethodPrivateKeyJWT: false,
		GrantTypeRefreshToken: true,
		SupportedScopes:       []string{"openid", "profile", "email", "address", "phone", "offline_access"},
	}

	provider, err := op.NewProvider(
		config,
		storage,
		op.StaticIssuer(issuer),
		op.WithAllowInsecure(), // Allow HTTP in development
	)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	// Ensure RFC 6749 5.1 cache-control headers on token and userinfo responses
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/oauth/token" || req.URL.Path == "/token" || req.URL.Path == "/userinfo" {
				w.Header().Set("Cache-Control", "no-store")
				w.Header().Set("Pragma", "no-cache")
			}
			next.ServeHTTP(w, req)
		})
	})

	// Mount the standard OIDC endpoints from zitadel/oidc
	r.Mount("/", provider)

	// Mount the login/consent UI
	r.Get("/login", LoginHandler(storage))
	r.Post("/login", LoginSubmitHandler(storage, provider))

	logger.Info("OIDC provider initialized", slog.String("issuer", issuer))

	return r, nil
}
