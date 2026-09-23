import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import client from '../api/client'
import type { User } from '../types'

interface AuthContextValue {
  user: User | null
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  loginWithGoogle: (credential: string) => Promise<void>
  updateUser: (user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

const TOKEN_KEY = 'devsync_token'
const USER_KEY = 'devsync_user'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? JSON.parse(raw) : null
  })

  // Refresh the stored user once per page load, so changes made server-side
  // (e.g. being promoted to admin) show up without logging out and back in.
  useEffect(() => {
    if (!localStorage.getItem(TOKEN_KEY)) return
    client
      .get<User>('/users/me')
      .then(({ data }) => setUser(data))
      .catch(() => {
        // offline or token expired; keep the cached user as before
      })
  }, [])

  useEffect(() => {
    if (user) {
      localStorage.setItem(USER_KEY, JSON.stringify(user))
    } else {
      localStorage.removeItem(USER_KEY)
    }
  }, [user])

  async function login(email: string, password: string) {
    const { data } = await client.post('/auth/login', { email, password })
    localStorage.setItem(TOKEN_KEY, data.token)
    setUser(data.user)
  }

  async function register(name: string, email: string, password: string) {
    const { data } = await client.post('/auth/register', { name, email, password })
    localStorage.setItem(TOKEN_KEY, data.token)
    setUser(data.user)
  }

  async function loginWithGoogle(credential: string) {
    const { data } = await client.post('/auth/google', { credential })
    localStorage.setItem(TOKEN_KEY, data.token)
    setUser(data.user)
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, login, register, loginWithGoogle, updateUser: setUser, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
