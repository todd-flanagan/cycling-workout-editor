package handlers

import (
	"context"
	"errors"
	"sync"

	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
	"golang.org/x/oauth2"
	"net/http"
)

// --- Mock UserRepository ---

type mockUserRepo struct {
	mu    sync.Mutex
	users map[string]*models.User
	// byGoogleID maps google_id -> user id for quick lookup.
	byGoogleID map[string]string
	createErr  error
	updateErr  error
	getErr     error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:      make(map[string]*models.User),
		byGoogleID: make(map[string]string),
	}
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) GetByGoogleID(_ context.Context, googleID string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	uid, ok := m.byGoogleID[googleID]
	if !ok {
		return nil, nil
	}
	u := m.users[uid]
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) Create(_ context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return m.createErr
	}
	cp := *user
	m.users[user.ID] = &cp
	m.byGoogleID[user.GoogleID] = user.ID
	return nil
}

func (m *mockUserRepo) Update(_ context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	cp := *user
	m.users[user.ID] = &cp
	return nil
}

// --- Mock FTPHistoryRepository ---

type mockFTPHistoryRepo struct {
	mu      sync.Mutex
	entries []models.FTPHistory
	listErr error
	createErr error
}

func newMockFTPHistoryRepo() *mockFTPHistoryRepo {
	return &mockFTPHistoryRepo{}
}

func (m *mockFTPHistoryRepo) List(_ context.Context, userID string) ([]models.FTPHistory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []models.FTPHistory
	for _, e := range m.entries {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockFTPHistoryRepo) Create(_ context.Context, entry *models.FTPHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return m.createErr
	}
	m.entries = append(m.entries, *entry)
	return nil
}

// --- Mock WorkoutRepository ---

type mockWorkoutRepo struct {
	mu       sync.Mutex
	workouts map[string]*models.Workout
	createErr error
	updateErr error
	getErr    error
}

func newMockWorkoutRepo() *mockWorkoutRepo {
	return &mockWorkoutRepo{
		workouts: make(map[string]*models.Workout),
	}
}

func (m *mockWorkoutRepo) GetByID(_ context.Context, id string, userID string) (*models.Workout, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	w, ok := m.workouts[id]
	if !ok || w.UserID != userID {
		return nil, nil
	}
	cp := *w
	return &cp, nil
}

func (m *mockWorkoutRepo) Create(_ context.Context, workout *models.Workout) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return m.createErr
	}
	cp := *workout
	m.workouts[workout.ID] = &cp
	return nil
}

func (m *mockWorkoutRepo) Update(_ context.Context, workout *models.Workout) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.workouts[workout.ID]; !ok {
		return errors.New("workout not found")
	}
	cp := *workout
	m.workouts[workout.ID] = &cp
	return nil
}

// --- Mock OAuthProvider ---

type mockOAuthProvider struct {
	authURL     string
	exchangeErr error
	token       *oauth2.Token
	client      *http.Client
}

func newMockOAuthProvider() *mockOAuthProvider {
	return &mockOAuthProvider{
		authURL: "https://accounts.google.com/o/oauth2/auth?mock=true",
		token:   &oauth2.Token{AccessToken: "mock-access-token"},
	}
}

func (m *mockOAuthProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return m.authURL
}

func (m *mockOAuthProvider) Exchange(_ context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	if m.exchangeErr != nil {
		return nil, m.exchangeErr
	}
	return m.token, nil
}

func (m *mockOAuthProvider) Client(_ context.Context, _ *oauth2.Token) *http.Client {
	if m.client != nil {
		return m.client
	}
	return http.DefaultClient
}
