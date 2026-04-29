import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  getToken,
  setToken,
  clearToken,
  apiFetch,
  apiGet,
  apiPost,
  apiPut,
  apiDelete,
} from './api';

describe('api', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
    delete window.location;
    window.location = { href: '' };
  });

  afterEach(() => {
    localStorage.clear();
  });

  describe('token management', () => {
    it('getToken returns null when no token', () => {
      expect(getToken()).toBeNull();
    });

    it('setToken stores token in localStorage', () => {
      setToken('my-token');
      expect(localStorage.getItem('cwe_auth_token')).toBe('my-token');
    });

    it('getToken returns stored token', () => {
      setToken('my-token');
      expect(getToken()).toBe('my-token');
    });

    it('clearToken removes token from localStorage', () => {
      setToken('my-token');
      clearToken();
      expect(getToken()).toBeNull();
    });
  });

  describe('apiFetch', () => {
    it('adds Authorization header when token exists', async () => {
      setToken('bearer-token');
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ data: 'test' }),
      });
      vi.stubGlobal('fetch', mockFetch);

      await apiFetch('/api/test');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/test',
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer bearer-token',
          }),
        })
      );
    });

    it('does not add Authorization header when no token', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      });
      vi.stubGlobal('fetch', mockFetch);

      await apiFetch('/api/test');

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.Authorization).toBeUndefined();
    });

    it('sets Content-Type to application/json', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      });
      vi.stubGlobal('fetch', mockFetch);

      await apiFetch('/api/test');

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers['Content-Type']).toBe('application/json');
    });

    it('redirects to /login on 401', async () => {
      setToken('expired-token');
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Unauthorized' }),
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(apiFetch('/api/test')).rejects.toThrow('Unauthorized');
      expect(window.location.href).toBe('/login');
      expect(getToken()).toBeNull();
    });

    it('throws error for non-OK responses', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: () => Promise.resolve({ error: 'Server error' }),
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(apiFetch('/api/test')).rejects.toThrow('Server error');
    });

    it('returns null for 204 responses', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 204,
      });
      vi.stubGlobal('fetch', mockFetch);

      const result = await apiFetch('/api/test');
      expect(result).toBeNull();
    });

    it('returns parsed JSON for successful responses', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ id: 1, name: 'test' }),
      });
      vi.stubGlobal('fetch', mockFetch);

      const result = await apiFetch('/api/test');
      expect(result).toEqual({ id: 1, name: 'test' });
    });

    it('handles non-JSON error responses', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: () => Promise.reject(new Error('not json')),
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(apiFetch('/api/test')).rejects.toThrow('Request failed: 500');
    });
  });

  describe('convenience methods', () => {
    let mockFetch;

    beforeEach(() => {
      mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ ok: true }),
      });
      vi.stubGlobal('fetch', mockFetch);
    });

    it('apiGet sends GET request', async () => {
      await apiGet('/api/users');
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/users',
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('apiPost sends POST request with body', async () => {
      await apiPost('/api/users', { name: 'Test' });
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/users',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ name: 'Test' }),
        })
      );
    });

    it('apiPut sends PUT request with body', async () => {
      await apiPut('/api/users/1', { name: 'Updated' });
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/users/1',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ name: 'Updated' }),
        })
      );
    });

    it('apiDelete sends DELETE request', async () => {
      await apiDelete('/api/users/1');
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/users/1',
        expect.objectContaining({ method: 'DELETE' })
      );
    });
  });
});
