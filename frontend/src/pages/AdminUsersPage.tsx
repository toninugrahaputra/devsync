import { useEffect, useState } from 'react'
import { Navigate } from 'react-router-dom'
import { Check, Copy, Search, UserPlus } from 'lucide-react'
import client from '../api/client'
import Avatar from '../components/Avatar'
import { useAuth } from '../context/AuthContext'
import type { AdminUser, Workspace } from '../types'

export default function AdminUsersPage() {
  const { user } = useAuth()

  // role is missing on sessions saved before it existed; AuthContext fills it
  // in from /users/me shortly after load, so wait instead of redirecting.
  if (!user?.role) {
    return (
      <div className="h-full overflow-y-auto bg-gray-50">
        <main className="max-w-5xl mx-auto p-6">
          <p className="text-gray-500 text-sm">Loading...</p>
        </main>
      </div>
    )
  }
  if (user.role !== 'admin') return <Navigate to="/" replace />

  return <AdminUsersList />
}

function AdminUsersList() {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [ownedWorkspaces, setOwnedWorkspaces] = useState<Workspace[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')

  // Per-row state, keyed by user id.
  const [selectedWorkspace, setSelectedWorkspace] = useState<Record<number, string>>({})
  const [addingFor, setAddingFor] = useState<number | null>(null)
  const [rowError, setRowError] = useState<Record<number, string>>({})
  const [copiedFor, setCopiedFor] = useState<number | null>(null)

  async function loadUsers(q: string) {
    setLoadError('')
    try {
      const { data } = await client.get<AdminUser[]>('/admin/users', { params: { q } })
      setUsers(data)
    } catch {
      setLoadError('Could not load users.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    // Only workspaces this admin owns can take new members.
    client.get<Workspace[]>('/workspaces').then(({ data }) => setOwnedWorkspaces(data.filter((w) => w.role === 'owner')))
  }, [])

  useEffect(() => {
    const t = setTimeout(() => loadUsers(query.trim()), 250)
    return () => clearTimeout(t)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query])

  async function addToWorkspace(u: AdminUser) {
    const workspaceId = selectedWorkspace[u.id]
    if (!workspaceId) return
    setAddingFor(u.id)
    setRowError((prev) => ({ ...prev, [u.id]: '' }))
    try {
      await client.post(`/workspaces/${workspaceId}/members`, { user_id: u.id })
      setSelectedWorkspace((prev) => ({ ...prev, [u.id]: '' }))
      await loadUsers(query.trim())
    } catch (err) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setRowError((prev) => ({ ...prev, [u.id]: msg || 'Could not add to workspace' }))
    } finally {
      setAddingFor(null)
    }
  }

  async function copyEmail(u: AdminUser) {
    try {
      await navigator.clipboard.writeText(u.email)
      setCopiedFor(u.id)
      setTimeout(() => setCopiedFor((id) => (id === u.id ? null : id)), 1500)
    } catch {
      // clipboard permission denied; the email is still visible to copy by hand
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-gray-50">
      <main className="max-w-5xl mx-auto p-6">
        <div className="flex flex-wrap items-end justify-between gap-3 mb-6">
          <div>
            <h1 className="text-xl font-semibold">Registered users</h1>
            <p className="text-sm text-gray-500">Everyone who has signed up, by email or with Google.</p>
          </div>
          <div className="relative w-full sm:w-72">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search name or email..."
              className="w-full border rounded pl-8 pr-3 py-2 text-sm bg-white"
            />
          </div>
        </div>

        {ownedWorkspaces.length === 0 && !loading && (
          <p className="text-sm text-gray-600 bg-white border rounded-lg p-3 mb-4">
            You don’t own any workspaces yet — create one first to add people to it.
          </p>
        )}

        {loadError && (
          <div className="text-sm text-red-600 flex items-center gap-3 mb-4">
            {loadError}
            <button onClick={() => loadUsers(query.trim())} className="underline hover:no-underline">
              Retry
            </button>
          </div>
        )}

        {loading ? (
          <p className="text-gray-500 text-sm">Loading...</p>
        ) : users.length === 0 ? (
          <p className="text-gray-500 text-sm">{query.trim() ? 'No users match your search.' : 'No users yet.'}</p>
        ) : (
          <>
            <p className="text-xs text-gray-400 mb-2">
              {users.length} user{users.length === 1 ? '' : 's'}
            </p>
            <div className="bg-white border rounded-lg divide-y">
              {users.map((u) => {
                const memberOf = new Set(u.workspaces.map((w) => w.id))
                const available = ownedWorkspaces.filter((w) => !memberOf.has(w.id))

                return (
                  <div key={u.id} className="flex flex-col md:flex-row md:items-center gap-3 px-4 py-3">
                    <div className="flex items-center gap-3 flex-1 min-w-0">
                      <Avatar name={u.name} avatarPath={u.avatar_path} size={36} />
                      <div className="min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <p className="text-sm font-medium text-gray-800 truncate">{u.name}</p>
                          <span
                            className={`text-[10px] uppercase tracking-wide rounded px-1.5 py-0.5 ${
                              u.auth_provider === 'google' ? 'bg-blue-50 text-blue-700' : 'bg-gray-100 text-gray-600'
                            }`}
                          >
                            {u.auth_provider === 'google' ? 'Google' : 'Email'}
                          </span>
                          {u.role === 'admin' && (
                            <span className="text-[10px] uppercase tracking-wide rounded px-1.5 py-0.5 bg-[#D6AE32]/20 text-[#8a6d17]">
                              Admin
                            </span>
                          )}
                        </div>
                        <div className="flex items-center gap-1.5">
                          <p className="text-xs text-gray-600 truncate">{u.email}</p>
                          <button
                            onClick={() => copyEmail(u)}
                            className="text-gray-400 hover:text-gray-700"
                            title="Copy email"
                          >
                            {copiedFor === u.id ? <Check size={12} className="text-green-600" /> : <Copy size={12} />}
                          </button>
                        </div>
                        <p className="text-xs text-gray-400">
                          Joined {new Date(u.created_at).toLocaleDateString(undefined, { day: 'numeric', month: 'short', year: 'numeric' })}
                          {' · '}
                          {u.workspaces.length === 0 ? 'no workspaces' : u.workspaces.map((w) => w.name).join(', ')}
                        </p>
                      </div>
                    </div>

                    {ownedWorkspaces.length > 0 && (
                      <div className="md:w-72 shrink-0">
                        {available.length === 0 ? (
                          <p className="text-xs text-gray-400 md:text-right">In all your workspaces</p>
                        ) : (
                          <div className="flex items-center gap-2">
                            <select
                              value={selectedWorkspace[u.id] ?? ''}
                              onChange={(e) => setSelectedWorkspace((prev) => ({ ...prev, [u.id]: e.target.value }))}
                              className="flex-1 min-w-0 border rounded px-2 py-1.5 text-sm bg-white"
                            >
                              <option value="">Add to workspace...</option>
                              {available.map((w) => (
                                <option key={w.id} value={w.id}>
                                  {w.name}
                                </option>
                              ))}
                            </select>
                            <button
                              onClick={() => addToWorkspace(u)}
                              disabled={!selectedWorkspace[u.id] || addingFor === u.id}
                              className="flex items-center gap-1 text-sm rounded px-3 py-1.5 bg-[#D6AE32] text-gray-900 font-medium hover:bg-[#B6942B] disabled:opacity-50"
                            >
                              <UserPlus size={14} />
                              {addingFor === u.id ? 'Adding...' : 'Add'}
                            </button>
                          </div>
                        )}
                        {rowError[u.id] && <p className="text-red-600 text-xs mt-1">{rowError[u.id]}</p>}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </>
        )}
      </main>
    </div>
  )
}
