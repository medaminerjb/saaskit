package middleware

import (
	"net/http"
)

// SecurityHeadersConfig allows customizing HTTP security headers.
type SecurityHeadersConfig struct {
	HSTSMaxAge            string
	ContentSecurityPolicy string
	FrameOptions          string
	ContentTypeOptions    string
	XSSProtection         string
	ReferrerPolicy        string
	PermissionsPolicy     string
}

// DefaultSecurityHeadersConfig provides industry-standard security headers defaults.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		HSTSMaxAge:            "max-age=31536000; includeSubDomains; preload",
		ContentSecurityPolicy: "default-src 'self'; frame-ancestors 'none';",
		FrameOptions:          "DENY",
		ContentTypeOptions:    "nosniff",
		XSSProtection:         "0",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
	}
}

// SecurityHeaders creates a middleware that injects security headers into every HTTP response.
func SecurityHeaders(cfg ...SecurityHeadersConfig) func(http.Handler) http.Handler {
	config := DefaultSecurityHeadersConfig()
	if len(cfg) > 0 {
		if cfg[0].HSTSMaxAge != "" {
			config.HSTSMaxAge = cfg[0].HSTSMaxAge
		}
		if cfg[0].ContentSecurityPolicy != "" {
			config.ContentSecurityPolicy = cfg[0].ContentSecurityPolicy
		}
		if cfg[0].FrameOptions != "" {
			config.FrameOptions = cfg[0].FrameOptions
		}
		if cfg[0].ContentTypeOptions != "" {
			config.ContentTypeOptions = cfg[0].ContentTypeOptions
		}
		if cfg[0].XSSProtection != "" {
			config.XSSProtection = cfg[0].XSSProtection
		}
		if cfg[0].ReferrerPolicy != "" {
			config.ReferrerPolicy = cfg[0].ReferrerPolicy
		}
		if cfg[0].PermissionsPolicy != "" {
			config.PermissionsPolicy = cfg[0].PermissionsPolicy
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" || r.URL.Scheme == "https"
			if isHTTPS && config.HSTSMaxAge != "" {
				w.Header().Set("Strict-Transport-Security", config.HSTSMaxAge)
			}
			w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
			w.Header().Set("X-Frame-Options", config.FrameOptions)
			w.Header().Set("X-XSS-Protection", config.XSSProtection)
			w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			w.Header().Set("Permissions-Policy", config.PermissionsPolicy)

			next.ServeHTTP(w, r)
		})
	}
}
