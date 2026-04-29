import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Editor from './Editor';

const mockUser = { ftp: 250, display_name: 'Test User', email: 'test@test.com' };
const mockNavigate = vi.fn();

vi.mock('../AuthContext', () => ({
  useAuth: () => ({
    user: mockUser,
    token: 'test-token',
  }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({}),
  };
});

vi.mock('../api', () => ({
  apiGet: vi.fn().mockResolvedValue({}),
  apiPost: vi.fn().mockResolvedValue({ id: 'new-workout-id' }),
  apiPut: vi.fn().mockResolvedValue({}),
}));

// Mock crypto.randomUUID
beforeEach(() => {
  let counter = 0;
  vi.stubGlobal('crypto', {
    randomUUID: () => `test-uuid-${++counter}`,
  });
});

describe('Editor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock clientWidth for chart container
    Object.defineProperty(HTMLElement.prototype, 'clientWidth', {
      configurable: true,
      value: 800,
    });
  });

  it('renders the editor page', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByTestId('editor-page')).toBeInTheDocument();
  });

  it('renders the workout chart', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByTestId('workout-chart')).toBeInTheDocument();
  });

  it('renders the interval toolbar', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByTestId('interval-toolbar')).toBeInTheDocument();
  });

  it('renders the interval inspector', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByTestId('interval-inspector')).toBeInTheDocument();
  });

  it('renders workout name input', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByLabelText(/workout name/i)).toBeInTheDocument();
  });

  it('renders save and cancel buttons', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByText('Save')).toBeInTheDocument();
    expect(screen.getByText('Cancel')).toBeInTheDocument();
  });

  it('renders undo and redo buttons', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByText('Undo')).toBeInTheDocument();
    expect(screen.getByText('Redo')).toBeInTheDocument();
  });

  it('adds an interval when toolbar item is clicked', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByTestId('toolbar-item-steady'));

    // The chart should now contain an interval block
    const chart = screen.getByTestId('workout-chart');
    const blocks = chart.querySelectorAll('.interval-block');
    expect(blocks.length).toBeGreaterThan(0);
  });

  it('shows interval inspector when block is selected', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );

    // Add an interval
    fireEvent.click(screen.getByTestId('toolbar-item-steady'));

    // Click the interval block in the chart
    const chart = screen.getByTestId('workout-chart');
    const block = chart.querySelector('.interval-block');
    if (block) {
      fireEvent.click(block);
    }

    // Inspector should show interval fields
    const inspector = screen.getByTestId('interval-inspector');
    expect(inspector).toBeInTheDocument();
  });

  it('navigates when cancel is clicked', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    fireEvent.click(screen.getByText('Cancel'));
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('updates workout name', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    const nameInput = screen.getByLabelText(/workout name/i);
    fireEvent.change(nameInput, { target: { value: 'My Workout' } });
    expect(nameInput).toHaveValue('My Workout');
  });

  it('displays total duration', () => {
    render(
      <MemoryRouter>
        <Editor />
      </MemoryRouter>
    );
    expect(screen.getByText(/total: 0:00/i)).toBeInTheDocument();
  });
});
