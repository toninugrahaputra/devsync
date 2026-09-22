import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import GoogleSignInButton from '../components/GoogleSignInButton'

export default function LoginPage() {
  const { login, loginWithGoogle } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const redirectTo = searchParams.get('redirect') || '/'
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  function errorMessage(err: unknown, fallback: string) {
    return (err as { response?: { data?: { error?: string } } })?.response?.data?.error || fallback
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      await login(email, password)
      navigate(redirectTo)
    } catch (err) {
      setError(errorMessage(err, 'Login failed'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleGoogle(credential: string) {
    setError('')
    try {
      await loginWithGoogle(credential)
      navigate(redirectTo)
    } catch (err) {
      setError(errorMessage(err, 'Google sign-in failed'))
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow-xl border border-gray-100 p-8 w-full max-w-sm space-y-4">
        <h1 className="text-2xl font-bold text-center text-gray-800">
          Dev<span className="text-[#D6AE32]">Sync</span>
        </h1>
        <p className="text-center text-gray-500 text-sm">Sign in to your boards</p>
        {error && <p className="text-red-600 text-sm text-center">{error}</p>}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#D6AE32]"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#D6AE32]"
          />
        </div>
        <button
          type="submit"
          disabled={submitting}
          className="w-full bg-[#D6AE32] text-gray-900 rounded py-2 font-medium hover:bg-[#B6942B] disabled:opacity-50"
        >
          {submitting ? 'Signing in...' : 'Sign in'}
        </button>

        <GoogleDivider />
        <GoogleSignInButton onCredential={handleGoogle} />

        <p className="text-center text-sm text-gray-500">
          No account?{' '}
          <Link to={`/register${redirectTo !== '/' ? `?redirect=${encodeURIComponent(redirectTo)}` : ''}`} className="text-[#B6942B] hover:underline">
            Register
          </Link>
        </p>
      </form>
    </div>
  )
}

function GoogleDivider() {
  if (!import.meta.env.VITE_GOOGLE_CLIENT_ID) return null
  return (
    <div className="flex items-center gap-2 text-xs text-gray-400">
      <div className="flex-1 h-px bg-gray-200" />
      or
      <div className="flex-1 h-px bg-gray-200" />
    </div>
  )
}
