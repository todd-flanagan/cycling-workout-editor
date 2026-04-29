import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import App from './App';

// Helper to control what useAuth returns
let mockAuthState = { token: null, user: null, loading: false, login: vi.fn(), logout: vi.fn() };

vi.mock('./AuthContext', () => ({
  useAuth: () => mockAuthState,
  AuthProvider: ({ children }) => <>{children}</>,
}));

// Mock pages to avoid deep rendering
vi.mock('./pages/Login', () => ({
  default: () => <div data-testid="login-page">Login Page</div>,
}));

vi.mock('./pages/Editor', () => ({
  default: () => <div data-testid="editor-page">Editor Page</div>,
}));

vi.mock('./pages/Profile', () => ({
  default: () => <div data-testid="profile-page">Profile Page</div>,
}));

describe('App routes', () => {
  it('renders login page at /login', () => {
    mockAuthState = { token: null, user: null, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/login']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('login-page')).toBeInTheDocument();
  });

  it('redirects to login when visiting /editor without token', () => {
    mockAuthState = { token: null, user: null, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/editor']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('login-page')).toBeInTheDocument();
  });

  it('renders editor page at /editor when authenticated', () => {
    mockAuthState = { token: 'valid-token', user: { id: '1' }, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/editor']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('editor-page')).toBeInTheDocument();
  });

  it('renders profile page at /profile when authenticated', () => {
    mockAuthState = { token: 'valid-token', user: { id: '1' }, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/profile']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('profile-page')).toBeInTheDocument();
  });

  it('redirects to login when visiting /profile without token', () => {
    mockAuthState = { token: null, user: null, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/profile']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('login-page')).toBeInTheDocument();
  });

  it('redirects / to /editor', () => {
    mockAuthState = { token: 'valid-token', user: { id: '1' }, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('editor-page')).toBeInTheDocument();
  });

  it('redirects unknown routes to /editor', () => {
    mockAuthState = { token: 'valid-token', user: { id: '1' }, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/nonexistent']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('editor-page')).toBeInTheDocument();
  });

  it('renders editor page at /editor/:id when authenticated', () => {
    mockAuthState = { token: 'valid-token', user: { id: '1' }, loading: false, login: vi.fn(), logout: vi.fn() };
    render(
      <MemoryRouter initialEntries={['/editor/abc-123']}>
        <App />
      </MemoryRouter>
    );
    expect(screen.getByTestId('editor-page')).toBeInTheDocument();
  });
});
