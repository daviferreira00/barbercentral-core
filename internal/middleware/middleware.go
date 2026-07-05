package middleware

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"

	"barbercentral-core/internal/shared"
)

type contextKey string

const (
	UserKey contextKey = "user"
)

var (
	ErrNotAuthenticated = errors.New("não autenticado")
)

type TokenUser struct {
	ID              string
	Email           string
	Role            string
	ClientID        string
	OriginalAdminID string
}

type AuthValidator interface {
	ValidateToken(tokenStr string) (*Claims, error)
}

type Claims struct {
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	ClientID        string `json:"client_id"`
	OriginalAdminID string `json:"original_admin_id,omitempty"`
	jwt.RegisteredClaims
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Dur("duration", time.Since(start)).
			Str("ip", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Msg("requisição HTTP processada")
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Error().
					Interface("panic", rvr).
					Str("stack", string(debug.Stack())).
					Msg("panic recuperada no handler HTTP")

				shared.RespondWithError(w, http.StatusInternalServerError, "Erro interno do servidor", nil)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func RequireAuth(validator AuthValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				shared.RespondWithError(w, http.StatusUnauthorized, "Cabeçalho de autorização ausente", nil)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				shared.RespondWithError(w, http.StatusUnauthorized, "Formato do cabeçalho de autorização inválido", nil)
				return
			}

			tokenStr := parts[1]
			claims, err := validator.ValidateToken(tokenStr)
			if err != nil {
				shared.RespondWithError(w, http.StatusUnauthorized, "Token inválido ou expirado", err)
				return
			}

			tokenUser := &TokenUser{
				ID:              claims.UserID,
				Email:           claims.Email,
				Role:            claims.Role,
				ClientID:        claims.ClientID,
				OriginalAdminID: claims.OriginalAdminID,
			}

			ctx := context.WithValue(r.Context(), UserKey, tokenUser)
			ctx = context.WithValue(ctx, "user_id", tokenUser.ID)
			ctx = context.WithValue(ctx, "user_email", tokenUser.Email)
			ctx = context.WithValue(ctx, "user_role", tokenUser.Role)
			ctx = context.WithValue(ctx, "client_id", tokenUser.ClientID)
			ctx = context.WithValue(ctx, "original_admin_id", tokenUser.OriginalAdminID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(ctx context.Context) (*TokenUser, error) {
	val := ctx.Value(UserKey)
	if val == nil {
		return nil, ErrNotAuthenticated
	}
	u, ok := val.(*TokenUser)
	if !ok {
		return nil, ErrNotAuthenticated
	}
	return u, nil
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenUser, err := GetUser(r.Context())
			if err != nil {
				shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
				return
			}

			allowed := false
			for _, role := range roles {
				if tokenUser.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				shared.RespondWithError(w, http.StatusForbidden, "Acesso negado: privilégios insuficientes", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleOrImpersonating libera o acesso tanto para quem tem um dos roles
// exigidos, quanto para uma sessão de admin impersonando (OriginalAdminID
// presente) — usado nas rotas que o seletor de barbearias do admin precisa
// continuar acessando mesmo depois de já ter entrado em uma barbearia (ex:
// listar barbearias de novo, ou impersonar outra, para "pular" de unidade).
func RequireRoleOrImpersonating(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenUser, err := GetUser(r.Context())
			if err != nil {
				shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
				return
			}

			if tokenUser.OriginalAdminID != "" {
				next.ServeHTTP(w, r)
				return
			}

			allowed := false
			for _, role := range roles {
				if tokenUser.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				shared.RespondWithError(w, http.StatusForbidden, "Acesso negado: privilégios insuficientes", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
