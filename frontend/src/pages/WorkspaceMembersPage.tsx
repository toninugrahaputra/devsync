import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, Copy, Mail, Trash2 } from 'lucide-react'
import client from '../api/client'
import Avatar from '../components/Avatar'
import { useConfirm } from '../context/ConfirmContext'
import type { InviteResponse, User, Workspace, WorkspaceMember } from '../types'

const EMAIL_RE = /^\S+@\S+\.\S+$/

export default function WorkspaceMembersPage() {
  const { workspaceId } = useParams()
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [members, setMembers] = useState<WorkspaceMember[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const confirm = useConfirm()

  const [memberQuery, setMemberQuery] = useState('')
  const [memberResults, setMemberResults] = useState<User[]>([])
  const [inviting, setInviting] = useState(false)
  const [inviteResult, setInviteResult] = useState<InviteResponse | null>(null)
  const [inviteError, setInviteError] = useState('')

  async function load() {
    setLoading(true)
    setLoadError('')
    try {
      const [w, m] = await Promise.all([
        client.get<Workspace>(`/workspaces/${workspaceId}`),
        client.get<WorkspaceMember[]>(`/workspaces/${workspaceId}/members`),
      ])
      setWorkspace(w.data)
      setMembers(m.data)
    } catch {
      setLoadError('Could not load this workspace — it may not exist, or you may not have access to it.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workspaceId])

  const isOwner = workspace?.role === 'owner'
  const looksLikeEmail = EMAIL_RE.test(memberQuery.trim())

  async function searchUsers(q: string) {
    setMemberQuery(q)
    setInviteResult(null)
    setInviteError('')
    if (!q.trim()) {
      setMemberResults([])
      return
    }
    const { data } = await client.get<User[]>('/users', { params: { q } })
    setMemberResults(data)
  }

  async function addMember(userId: number) {
    await client.post(`/workspaces/${workspaceId}/members`, { user_id: userId })
    setMemberQuery('')
    setMemberResults([])
    load()
  }

  async function removeMember(userId: number) {
    const member = members.find((m) => m.user_id === userId)
    const ok = await confirm({
      title: 'Remove member',
      message: `Remove ${member?.name ?? 'this member'} from the workspace?`,
      confirmLabel: 'Remove',
      danger: true,
    })
    if (!ok) return
    await client.delete(`/workspaces/${workspaceId}/members/${userId}`)
    load()
  }

  async function inviteByEmail() {
    const email = memberQuery.trim()
    if (!email) return
    setInviting(true)
    setInviteResult(null)
    setInviteError('')
    try {
      const { data } = await client.post<InviteResponse>(`/workspaces/${workspaceId}/invites`, { email })
      setInviteResult(data)
      setMemberQuery('')
      setMemberResults([])
      if (data.already_member) load()
    } catch (err) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setInviteError(msg || 'Could not send invite')
    } finally {
      setInviting(false)
    }
  }

  async function copyLink(link: string) {
    try {
      await navigator.clipboard.writeText(link)
    } catch {
      // clipboard permission denied; the link is still shown for manual copy
    }
  }

  if (loading) {
    return (
      <div className="h-full overflow-y-auto bg-gray-50">
        <main className="max-w-3xl mx-auto p-6">
          <p className="text-gray-500 text-sm">Loading...</p>
        </main>
      </div>
    )
  }

  if (loadError) {
    return (
      <div className="h-full overflow-y-auto bg-gray-50">
        <main className="max-w-3xl mx-auto p-6">
          <Link to="/" className="text-sm text-[#B6942B] hover:underline">
            ← Workspaces
          </Link>
          <div className="mt-4 text-sm text-red-600 flex items-center gap-3">
            {loadError}
            <button onClick={load} className="underline hover:no-underline">
              Retry
            </button>
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto bg-gray-50">
      <main className="max-w-3xl mx-auto p-6">
        <Link
          to={`/w/${workspaceId}`}
          className="flex items-center gap-1 text-sm text-[#B6942B] hover:underline w-fit"
        >
          <ArrowLeft size={14} /> {workspace?.name}
        </Link>
        <h1 className="text-xl font-semibold mt-2 mb-6">Members</h1>

        {isOwner && (
          <div className="bg-white border rounded-lg p-4 mb-6">
            <h2 className="text-sm font-semibold text-gray-700 mb-2">Invite people</h2>
            <div className="relative mb-3">
              <input
                value={memberQuery}
                onChange={(e) => searchUsers(e.target.value)}
                placeholder="Search by name/email, or type an email to invite..."
                className="w-full border rounded px-3 py-2 text-sm"
              />
              {memberResults.length > 0 && (
                <div className="absolute z-10 bg-white border rounded shadow mt-1 w-full max-h-48 overflow-auto">
                  {memberResults.map((u) => (
                    <button
                      key={u.id}
                      onClick={() => addMember(u.id)}
                      className="flex items-center gap-2 w-full text-left px-3 py-2 text-sm hover:bg-gray-100"
                    >
                      <Avatar name={u.name} avatarPath={u.avatar_path} size={22} />
                      {u.name} <span className="text-gray-400">({u.email})</span>
                    </button>
                  ))}
                </div>
              )}
            </div>

            {looksLikeEmail && memberResults.length === 0 && (
              <button
                onClick={inviteByEmail}
                disabled={inviting}
                className="flex items-center gap-1.5 text-sm border rounded px-3 py-1.5 mb-3 hover:bg-gray-50 disabled:opacity-50"
              >
                <Mail size={14} />
                {inviting ? 'Inviting...' : `Invite ${memberQuery.trim()} by email`}
              </button>
            )}

            {inviteError && <p className="text-red-600 text-xs mb-3">{inviteError}</p>}

            {inviteResult && (
              <div className="bg-gray-50 border rounded p-3 text-sm">
                {inviteResult.already_member ? (
                  <p className="text-gray-700">This person already has an account — added to the workspace directly.</p>
                ) : (
                  <>
                    <p className="text-gray-700 mb-1.5">
                      {inviteResult.email_sent
                        ? 'Invite email sent. They can also use this link:'
                        : "Email isn’t configured on the server — share this link with them manually:"}
                    </p>
                    <div className="flex items-center gap-2">
                      <input
                        readOnly
                        value={inviteResult.invite_link}
                        className="flex-1 border rounded px-2 py-1 text-xs bg-white"
                        onFocus={(e) => e.target.select()}
                      />
                      <button
                        onClick={() => inviteResult.invite_link && copyLink(inviteResult.invite_link)}
                        className="flex items-center gap-1 text-xs border rounded px-2 py-1 hover:bg-white"
                      >
                        <Copy size={12} /> Copy
                      </button>
                    </div>
                  </>
                )}
              </div>
            )}
          </div>
        )}

        <div className="bg-white border rounded-lg divide-y">
          {members.map((m) => (
            <div key={m.user_id} className="flex items-center justify-between px-4 py-3">
              <div className="flex items-center gap-3">
                <Avatar name={m.name} avatarPath={m.avatar_path} size={32} />
                <div>
                  <p className="text-sm font-medium text-gray-800">{m.name}</p>
                  <p className="text-xs text-gray-500">{m.email}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-xs text-gray-400 uppercase">{m.role}</span>
                {isOwner && m.role !== 'owner' && (
                  <button
                    onClick={() => removeMember(m.user_id)}
                    className="flex items-center gap-1 text-red-500 text-xs hover:underline"
                  >
                    <Trash2 size={12} /> Remove
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  )
}
