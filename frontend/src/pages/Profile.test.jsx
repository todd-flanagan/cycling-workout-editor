import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Profile from './Profile';

const mockUser = {
  display_name: 'John Doe',
  email: 'john@example.com',
  ftp: 250,
};

const mockNavigate = vi.fn();
const mockLogout = vi.fn();

vi.mock('../AuthContext', () => ({
  useAuth: () => ({
    user: mockUser,
    logout: mockLogout,
  }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../api', () => ({
  apiGet: vi.fn().mockResolvedValue([]),
  apiPost: vi.fn().mockResolvedValue({}),
  apiPut: vi.fn().mockResolvedValue({}),
}));

describe('Profile', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders user profile information', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    expect(screen.getByText('User Profile')).toBeInTheDocument();
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('john@example.com')).toBeInTheDocument();
    expect(screen.getByText('250W')).toBeInTheDocument();
  });

  it('renders FTP input form', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    expect(screen.getByLabelText(/ftp.*watts/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/date/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/source/i)).toBeInTheDocument();
    expect(screen.getByText('Save FTP')).toBeInTheDocument();
  });

  it('renders ramp test calculator', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    expect(screen.getByLabelText(/max 1-minute power/i)).toBeInTheDocument();
    expect(screen.getByText('Calculate')).toBeInTheDocument();
  });

  it('calculates FTP from ramp test', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    const rampInput = screen.getByLabelText(/max 1-minute power/i);
    fireEvent.change(rampInput, { target: { value: '300' } });
    fireEvent.click(screen.getByText('Calculate'));

    expect(screen.getByText(/225W/)).toBeInTheDocument();
    expect(screen.getByText(/75% of 300W/)).toBeInTheDocument();
    expect(screen.getByText('Confirm & Save')).toBeInTheDocument();
  });

  it('renders FTP history section', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    expect(screen.getByText('FTP History')).toBeInTheDocument();
  });

  it('shows empty state when no FTP history', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    expect(screen.getByText(/no ftp history/i)).toBeInTheDocument();
  });

  it('renders logout button', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    expect(screen.getByText('Log Out')).toBeInTheDocument();
  });

  it('calls logout and navigates on logout button click', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    fireEvent.click(screen.getByText('Log Out'));
    expect(mockLogout).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });

  it('shows error for invalid FTP value', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    const ftpInput = screen.getByLabelText(/ftp.*watts/i);
    fireEvent.change(ftpInput, { target: { value: '10' } });
    fireEvent.submit(screen.getByText('Save FTP').closest('form'));
    expect(screen.getByText(/ftp must be between 50 and 500/i)).toBeInTheDocument();
  });

  it('shows error for invalid ramp test value', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    const rampInput = screen.getByLabelText(/max 1-minute power/i);
    fireEvent.change(rampInput, { target: { value: '10' } });
    fireEvent.click(screen.getByText('Calculate'));
    expect(screen.getByText(/must be between 50 and 1000/i)).toBeInTheDocument();
  });

  it('navigates to editor on editor button click', () => {
    render(
      <MemoryRouter>
        <Profile />
      </MemoryRouter>
    );
    fireEvent.click(screen.getByText('Editor'));
    expect(mockNavigate).toHaveBeenCalledWith('/editor');
  });
});
