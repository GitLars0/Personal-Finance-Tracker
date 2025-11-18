import { useCallback } from 'react';

export default function useAuthFetch() {
  const token = localStorage.getItem('token');

  const authFetch = useCallback(async (url, opts = {}) => {
    const headers = opts.headers || {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    if (opts.body && !headers['Content-Type']) headers['Content-Type'] = 'application/json';

    const response = await fetch(url, { ...opts, headers });
    if (!response.ok) {
      let err = new Error('Request failed');
      try {
        const data = await response.json();
        err.message = data.error || data.message || err.message;
      } catch (e) {}
      err.status = response.status;
      throw err;
    }
    return response.json();
  }, [token]);

  return authFetch;
}
