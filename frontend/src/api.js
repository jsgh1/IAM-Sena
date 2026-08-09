const API_BASE = import.meta.env.VITE_API_URL || '/api'
const ACCESS_KEY = 'iam_access_token'
const REFRESH_KEY = 'iam_refresh_token'

export const authStore = {
  access: () => localStorage.getItem(ACCESS_KEY),
  refresh: () => localStorage.getItem(REFRESH_KEY),
  set(access, refresh) {
    if (access) localStorage.setItem(ACCESS_KEY, access)
    if (refresh) localStorage.setItem(REFRESH_KEY, refresh)
  },
  clear() {
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  }
}

async function refreshAccess() {
  const refresh = authStore.refresh()
  if (!refresh) return false
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refresh })
  })
  if (!res.ok) { authStore.clear(); return false }
  const data = await res.json(); authStore.set(data.access_token); return true
}

export async function api(path, options = {}, retry = true) {
  const headers = { ...(options.body ? { 'Content-Type': 'application/json' } : {}), ...(options.headers || {}) }
  const token = authStore.access(); if (token) headers.Authorization = `Bearer ${token}`
  let res = await fetch(`${API_BASE}${path}`, { ...options, headers })
  if (res.status === 401 && retry && !path.startsWith('/auth/')) {
    if (await refreshAccess()) return api(path, options, false)
  }
  const text = await res.text(); let data = null
  try { data = text ? JSON.parse(text) : null } catch { data = { error: text || 'Respuesta inválida' } }
  if (!res.ok) throw new Error(data?.error || `Error HTTP ${res.status}`)
  return data
}

export async function login(email, password) {
  const data = await api('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) })
  authStore.set(data.access_token, data.refresh_token)
  return data.user
}

export async function logout() {
  const refresh = authStore.refresh()
  try { if (refresh) await api('/auth/logout', { method: 'POST', body: JSON.stringify({ refresh_token: refresh }) }, false) } finally { authStore.clear() }
}
