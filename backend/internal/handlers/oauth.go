package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
	"golang.org/x/oauth2"
)

// OAuthProvider abstracts the OAuth2 flow for testability.
type OAuthProvider interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}

// GoogleUserInfo represents the response from Google's userinfo endpoint.
type GoogleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserInfoFetcher is a function type for fetching user info from an OAuth provider.
// This makes the handler testable by allowing injection of a mock fetcher.
type UserInfoFetcher func(ctx context.Context, client *http.Client) (*GoogleUserInfo, error)

// DefaultUserInfoFetcher fetches user info from Google's userinfo API.
func DefaultUserInfoFetcher(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read user info body: %w", err)
	}

	var info GoogleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse user info: %w", err)
	}
	return &info, nil
}

// SetUserInfoFetcher sets a custom user info fetcher (for testing).
func (h *Handler) SetUserInfoFetcher(f UserInfoFetcher) {
	h.userInfoFetcher = f
}

// getUserInfoFetcher returns the configured fetcher or the default.
func (h *Handler) getUserInfoFetcher() UserInfoFetcher {
	if h.userInfoFetcher != nil {
		return h.userInfoFetcher
	}
	return DefaultUserInfoFetcher
}

// userInfoFetcher is an optional override for testing.
// It is not exported; use SetUserInfoFetcher to configure.
// We store it on the handler struct below via a field added in the init.
// (This field is declared in handlers.go but we reference it here.)

// GoogleLogin redirects the user to Google's OAuth consent screen.
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// In production you would use a random state parameter and store it
	// in a cookie or session to prevent CSRF. For simplicity we use a
	// fixed string here; a real implementation should generate and validate state.
	url := h.OAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback from Google.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		respondError(w, http.StatusBadRequest, "missing code parameter")
		return
	}

	token, err := h.OAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to exchange token")
		return
	}

	client := h.OAuthConfig.Client(r.Context(), token)
	fetcher := h.getUserInfoFetcher()
	userInfo, err := fetcher(r.Context(), client)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	// Upsert user.
	user, err := h.UserRepo.GetByGoogleID(r.Context(), userInfo.Sub)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}

	if user == nil {
		user = &models.User{
			ID:          uuid.New().String(),
			GoogleID:    userInfo.Sub,
			Email:       userInfo.Email,
			DisplayName: userInfo.Name,
			FTP:         200, // Default FTP for new users.
		}
		if err := h.UserRepo.Create(r.Context(), user); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	} else {
		// Update email/name in case they changed.
		user.Email = userInfo.Email
		user.DisplayName = userInfo.Name
		if err := h.UserRepo.Update(r.Context(), user); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
	}

	// Generate JWT.
	jwtToken, err := h.JWTManager.GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Redirect to frontend with token.
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", h.FrontendURL, jwtToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
