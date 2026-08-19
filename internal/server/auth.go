package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/podcctv/detective-chicken/internal/model"
	"github.com/podcctv/detective-chicken/internal/store"
)

const sessionCookie = "dc_session"

type principal struct {
	User     model.User
	TenantID string
	Token    string
}

type principalKey struct{}

func requestPrincipal(r *http.Request) principal {
	value, _ := r.Context().Value(principalKey{}).(principal)
	return value
}

func sessionToken(r *http.Request) string {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		return cookie.Value
	}
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

func (a *API) optionalPrincipal(r *http.Request) (principal, bool) {
	token := sessionToken(r)
	user, tenantID, err := a.store.UserBySession(token)
	if err != nil {
		return principal{}, false
	}
	return principal{User: user, TenantID: tenantID, Token: token}, true
}

func (a *API) authenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := a.optionalPrincipal(r)
		if !ok {
			apiError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "login required")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, p)))
	}
}

func (a *API) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return a.authenticated(func(w http.ResponseWriter, r *http.Request) {
		if requestPrincipal(r).User.Role != "admin" {
			apiError(w, http.StatusForbidden, "ADMIN_REQUIRED", "administrator role required")
			return
		}
		next(w, r)
	})
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expires time.Time) {
	secure := r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") || strings.EqualFold(os.Getenv("DETECTIVE_CHICKEN_COOKIE_SECURE"), "true")
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", Expires: expires, MaxAge: int(time.Until(expires).Seconds()), HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	setSessionCookie(w, r, "", time.Unix(1, 0))
}

func (a *API) authStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{"settings": a.store.PublicSettings(), "authenticated": false}
	if p, ok := a.optionalPrincipal(r); ok {
		response["authenticated"] = true
		response["user"] = p.User
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *API) registerUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password"`
	}
	if err := decode(r, &input, 32<<10); err != nil {
		apiError(w, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	user, err := a.store.RegisterUser(input.Username, input.DisplayName, input.Password)
	if err != nil {
		status, code := http.StatusBadRequest, "INVALID_ACCOUNT"
		if errors.Is(err, store.ErrConflict) {
			status, code = http.StatusConflict, "USERNAME_TAKEN"
		} else if errors.Is(err, store.ErrRegistrationClosed) {
			status, code = http.StatusForbidden, "REGISTRATION_CLOSED"
		}
		apiError(w, status, code, err.Error())
		return
	}
	_, token, expires, err := a.store.CreateSession(input.Username, input.Password)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "SESSION_FAILED", "account created but session could not be started")
		return
	}
	setSessionCookie(w, r, token, expires)
	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "expires_at": expires})
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var input struct{ Username, Password string }
	if err := decode(r, &input, 32<<10); err != nil {
		apiError(w, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	user, token, expires, err := a.store.CreateSession(input.Username, input.Password)
	if err != nil {
		apiError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "username or password is incorrect")
		return
	}
	setSessionCookie(w, r, token, expires)
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "expires_at": expires})
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	a.store.DeleteSession(requestPrincipal(r).Token)
	clearSessionCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) changePassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := decode(r, &input, 32<<10); err != nil {
		apiError(w, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	if err := a.store.ChangePassword(requestPrincipal(r).User.ID, input.CurrentPassword, input.NewPassword); err != nil {
		apiError(w, http.StatusBadRequest, "PASSWORD_CHANGE_FAILED", err.Error())
		return
	}
	clearSessionCookie(w, r)
	writeJSON(w, http.StatusOK, map[string]any{"changed": true, "login_required": true})
}

func (a *API) completePasswordReset(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := decode(r, &input, 32<<10); err != nil {
		apiError(w, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	if err := a.store.CompletePasswordReset(input.Token, input.NewPassword); err != nil {
		apiError(w, http.StatusBadRequest, "RESET_FAILED", "reset token is invalid, expired, or the password is not acceptable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"changed": true})
}

func (a *API) users(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": a.store.Users()})
}

func (a *API) updateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Role string `json:"role"`
	}
	if err := decode(r, &input, 32<<10); err != nil {
		apiError(w, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	user, err := a.store.UpdateUserRole(requestPrincipal(r).User.ID, r.PathValue("id"), input.Role)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, store.ErrForbidden) {
			status = http.StatusForbidden
		}
		apiError(w, status, "ROLE_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (a *API) createPasswordReset(w http.ResponseWriter, r *http.Request) {
	token, expires, err := a.store.CreatePasswordReset(r.PathValue("id"))
	if err != nil {
		apiError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "expires_at": expires})
}

func (a *API) settings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.PublicSettings())
}

func (a *API) updateSettings(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RegistrationEnabled bool `json:"registration_enabled"`
	}
	if err := decode(r, &input, 32<<10); err != nil {
		apiError(w, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	a.store.SetRegistrationEnabled(input.RegistrationEnabled)
	writeJSON(w, http.StatusOK, a.store.PublicSettings())
}
