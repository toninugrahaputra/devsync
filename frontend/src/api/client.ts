import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

const client = axios.create({
  baseURL: API_BASE_URL,
})

// Uploaded files (attachments, avatars) are served as static files off the
// backend's origin, not under /api/v1 — this strips that suffix to get there.
const API_ORIGIN = API_BASE_URL.replace(/\/api\/v1\/?$/, '')

export function fileUrl(path: string): string {
  return `${API_ORIGIN}/${path.replace(/^\/+/, '')}`
}

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('devsync_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export default client
