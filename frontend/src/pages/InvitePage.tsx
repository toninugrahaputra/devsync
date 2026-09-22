import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import client from '../api/client'
import { useAuth } from '../context/AuthContext'
import type { InvitePreview } from '../types'

export default function InvitePage() {
  const { token } = useParams()
  const { user } = useAuth()
  const navigate = useNavigate()
  const [preview, setPreview] = useState<InvitePreview | null>(null)
  const [error, setError] = useState('')
  const [accepting, setAccepting] = useState(false)

  useEffect(() => {
    if (!token) return
    client
      .get<InvitePreview>(`/invites/${token}`)
      .then(({ data }) => setPreview(data))
      .catch(() => setError('This invite link is invalid or has expired.'))
  }, [token])

  useEffect(() => {
    if (!preview || !user || !token || preview.already_accepted) return
    if (user.email.toLowerCase() !== preview.email.toLowerCase()) return

    setAccepting(true)
    client
      .post(`/invites/${token}/accept`)
      .then(() => navigate('/'))
      .catch((err) => {
        setError(err?.response?.data?.error || 'Could not accept this invite.')
        setAccepting(false)
      })
  }, [preview, user, token, navigate])

  const redirect = `/invite/${token}`

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white rounded-lg shadow-xl border border-gray-100 p-8 w-full max-w-sm text-center space-y-4">
        <h1 className="text-xl font-bold text-gray-800">
          Dev<span className="text-[#D6AE32]">Sync</span>
        </h1>

        {error && <p className="text-red-600 text-sm">{error}</p>}

        {!error && !preview && <p className="text-gray-500 text-sm">Loading invite...</p>}

        {!error && preview && preview.already_accepted && (
          <p className="text-gray-600 text-sm">This invite has already been used.</p>
        )}

        {!error && preview && !preview.already_accepted && (
          <>
            <p className="text-gray-700 text-sm">
              You've been invited to join <span className="font-semibold">{preview.workspace_name}</span> as{' '}
              <span className="font-medium">{preview.email}</span>.
            </p>

            {accepting && <p className="text-gray-500 text-sm">Joining workspace...</p>}

            {!accepting && user && user.email.toLowerCase() !== preview.email.toLowerCase() && (
              <p className="text-sm text-red-600">
                You're signed in as {user.email}, but this invite was sent to {preview.email}. Log out and sign in with
                the right account to accept it.
              </p>
            )}

            {!accepting && !user && (
              <div className="flex flex-col gap-2">
                <Link
                  to={`/login?redirect=${encodeURIComponent(redirect)}`}
                  className="w-full bg-[#D6AE32] text-gray-900 rounded py-2 font-medium hover:bg-[#B6942B]"
                >
                  Sign in
                </Link>
                <Link
                  to={`/register?redirect=${encodeURIComponent(redirect)}`}
                  className="w-full border rounded py-2 font-medium text-gray-700 hover:bg-gray-50"
                >
                  Create account
                </Link>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
