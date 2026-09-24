// REACT_APP_API_URL is empty when the API is served from the same origin (the
// ingress routes /api to the backend, and CRA's dev proxy does the same).
const BASE = process.env.REACT_APP_API_URL || '';
const STORAGE_KEY = 'room-chat.session';
// Refresh a little early so a request never goes out with a token that
// expires in flight.
const SKEW_MS = 30 * 1000;

export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

let session = load();
let refreshing = null;
const listeners = new Set();

function load() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY)) || null;
  } catch {
    return null;
  }
}

function save(next) {
  session = next;
  try {
    if (next) localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    else localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Storage blocked: the session just won't survive a reload.
  }
  listeners.forEach((fn) => fn(next));
}

function fromTokens(t) {
  return {
    accessToken: t.access_token,
    refreshToken: t.refresh_token,
    expiresAt: Date.now() + t.expires_in * 1000,
  };
}

// Called whenever the session is set or cleared (login, refresh, logout).
export function onSessionChange(fn) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export const hasSession = () => !!session;

async function raw(path, { method = 'GET', body, token } = {}) {
  let res;
  try {
    res = await fetch(BASE + path, {
      method,
      headers: {
        ...(body ? { 'Content-Type': 'application/json' } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch {
    throw new ApiError(0, 'Cannot reach the server');
  }
  if (res.status === 204) return null;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new ApiError(res.status, data.error || `Request failed (${res.status})`);
  return data;
}

// One refresh at a time: several requests can hit an expired token together,
// and user-service rotates the refresh token on use.
function refresh() {
  if (!session) return Promise.reject(new ApiError(401, 'Not signed in'));
  if (!refreshing) {
    const { refreshToken } = session;
    refreshing = raw('/api/auth/refresh', { method: 'POST', body: { refresh_token: refreshToken } })
      .then((t) => {
        save(fromTokens(t));
        return session.accessToken;
      })
      .catch((err) => {
        // Only a definitive rejection ends the session; a network blip or a
        // 502 from user-service should not sign everyone out.
        if (err.status === 401) save(null);
        throw err;
      })
      .finally(() => {
        refreshing = null;
      });
  }
  return refreshing;
}

// A valid access token, refreshing it first if it is about to expire.
export async function accessToken() {
  if (!session) throw new ApiError(401, 'Not signed in');
  if (Date.now() < session.expiresAt - SKEW_MS) return session.accessToken;
  return refresh();
}

async function request(path, opts) {
  let token = await accessToken();
  try {
    return await raw(path, { ...opts, token });
  } catch (err) {
    if (err.status !== 401) throw err;
    token = await refresh(); // revoked or expired early: one retry
    return raw(path, { ...opts, token });
  }
}

export async function login(username, password) {
  const t = await raw('/api/auth/login', { method: 'POST', body: { username, password } });
  save(fromTokens(t));
}

export function logout() {
  save(null);
}

export const me = () => request('/api/me');
export const listRooms = () => request('/api/rooms');
export const createRoom = (name) => request('/api/rooms', { method: 'POST', body: { name } });
export const getRoom = (id) => request(`/api/rooms/${id}`);
export const listMembers = (id) => request(`/api/rooms/${id}/members`);
export const addMember = (id, username) =>
  request(`/api/rooms/${id}/members`, { method: 'POST', body: { username } });
export const removeMember = (id, username) =>
  request(`/api/rooms/${id}/members/${encodeURIComponent(username)}`, { method: 'DELETE' });
export const listMessages = (id, { before, limit } = {}) => {
  const q = new URLSearchParams();
  if (before) q.set('before', before);
  if (limit) q.set('limit', limit);
  const qs = q.toString();
  return request(`/api/rooms/${id}/messages${qs ? `?${qs}` : ''}`);
};

export function socketUrl(roomId, token) {
  const base = BASE || window.location.origin;
  const url = new URL(`/api/rooms/${roomId}/ws`, base);
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
  url.searchParams.set('token', token);
  return url.toString();
}
