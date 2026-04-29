import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Login from './Login';

// Mock AuthContext
const mockLogin = vi.fn();
const mockNavigate = vi.fn();
let mockToken = null;

vi.mock('../AuthContext', () => ({
  useAuth: () => ({
    login: mockLogin,
    token: mockToken,
  }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useSearchParams: () => [new URLSearchParams()],
  };
});

describe('Login', () => {
  beforeEach(() => {
    mockToken = null;
    vi.clearAllMocks();
    // Mock window.location
    delete window.location;
    window.location = { href: '' };
  });

  it('renders the login page', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );
    expect(screen.getByText('Cycling Workout Editor')).toBeInTheDocument();
  });

  it('renders the Sign in with Google button', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );
    expect(screen.getByText('Sign in with Google')).toBeInTheDocument();
  });

  it('renders the subtitle', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );
    expect(
      screen.getByText(/build, visualize, and export/i)
    ).toBeInTheDocument();
  });

  it('redirects to Google OAuth on button click', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );
    fireEvent.click(screen.getByText('Sign in with Google'));
    expect(window.location.href).toBe('/auth/google/login');
  });

  it('redirects to editor if already logged in', () => {
    mockToken = 'existing-token';
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );
    expect(mockNavigate).toHaveBeenCalledWith('/editor', { replace: true });
  });
});
