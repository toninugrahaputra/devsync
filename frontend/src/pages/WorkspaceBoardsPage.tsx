import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Plus, Settings, Trash2, Users } from 'lucide-react'
import client from '../api/client'
import ArchivedCardsModal from '../components/ArchivedCardsModal'
import BoardSettingsMenu from '../components/BoardSettingsMenu'
import { useAuth } from '../context/AuthContext'
import { useConfirm } from '../context/ConfirmContext'
import type { Board, BoardDetail, Workspace, WorkspaceMember } from '../types'

export default function WorkspaceBoardsPage() {
  const { workspaceId } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const confirm = useConfirm()
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [boards, setBoards] = useState<Board[]>([])
  const [boardDetails, setBoardDetails] = useState<Record<number, BoardDetail>>({})
  const [settingsBoardId, setSettingsBoardId] = useState<number | null>(null)
  const [settingsAnchor, setSettingsAnchor] = useState<{ x: number; y: number } | null>(null)
  const [archivedBoardId, setArchivedBoardId] = useState<number | null>(null)
  const [members, setMembers] = useState<WorkspaceMember[]>([])
  const [newBoardName, setNewBoardName] = useState('')
  const [editingName, setEditingName] = useState(false)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')

  async function load() {
    setLoading(true)
    setLoadError('')
    try {
      const [w, b, m] = await Promise.all([
        client.get<Workspace>(`/workspaces/${workspaceId}`),
        client.get<Board[]>(`/workspaces/${workspaceId}/boards`),
        client.get<WorkspaceMember[]>(`/workspaces/${workspaceId}/members`),
      ])
      setWorkspace(w.data)
      setBoards(b.data)
      setMembers(m.data)
    } catch {
      setLoadError("Could not load this workspace — it may not exist, or you may not have access to it.")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workspaceId])

  // Fetched per-board (not returned by the plain boards list) purely to know
  // which boards the current user is admin on, so the edit icon only shows
  // where it'll actually work — and to feed BoardSettingsMenu, which needs
  // the full member list.
  useEffect(() => {
    if (boards.length === 0) {
      setBoardDetails({})
      return
    }
    Promise.all(
      boards.map((b) =>
        client.get<BoardDetail>(`/boards/${b.id}`).then(
          (r) => [b.id, r.data] as const,
          () => null
        )
      )
    ).then((entries) => {
      const map: Record<number, BoardDetail> = {}
      for (const entry of entries) {
        if (entry) map[entry[0]] = entry[1]
      }
      setBoardDetails(map)
    })
  }, [boards])

  async function refreshBoardDetail(boardId: number) {
    const { data } = await client.get<BoardDetail>(`/boards/${boardId}`)
    setBoardDetails((prev) => ({ ...prev, [boardId]: data }))
  }

  const isOwner = workspace?.role === 'owner'

  async function renameWorkspace(name: string) {
    setEditingName(false)
    if (!workspace || !name.trim() || name === workspace.name) return
    await client.put(`/workspaces/${workspaceId}`, { name })
    load()
  }

  async function deleteWorkspace() {
    if (!workspace) return
    const ok = await confirm({
      title: 'Delete workspace',
      message: `Delete "${workspace.name}"? This removes every board inside it. This can't be undone.`,
      confirmLabel: 'Delete',
      danger: true,
    })
    if (!ok) return
    await client.delete(`/workspaces/${workspaceId}`)
    navigate('/')
  }

  async function handleCreateBoard(e: FormEvent) {
    e.preventDefault()
    if (!newBoardName.trim()) return
    await client.post(`/workspaces/${workspaceId}/boards`, { name: newBoardName })
    setNewBoardName('')
    load()
  }

  if (loading) {
    return (
      <div className="h-full overflow-y-auto bg-gray-50">
        <main className="max-w-5xl mx-auto p-6">
          <p className="text-gray-500 text-sm">Loading...</p>
        </main>
      </div>
    )
  }

  if (loadError) {
    return (
      <div className="h-full overflow-y-auto bg-gray-50">
        <main className="max-w-5xl mx-auto p-6">
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
      <main className="max-w-5xl mx-auto p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <Link to="/" className="text-sm text-[#B6942B] hover:underline">
              ← Workspaces
            </Link>
            {editingName ? (
              <input
                autoFocus
                defaultValue={workspace?.name}
                onBlur={(e) => renameWorkspace(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && (e.target as HTMLInputElement).blur()}
                className="block text-xl font-semibold outline-none border-b border-[#D6AE32]"
              />
            ) : (
              <h1
                onClick={() => isOwner && setEditingName(true)}
                className={`text-xl font-semibold ${isOwner ? 'block w-fit cursor-text hover:bg-gray-100 rounded px-1 -mx-1' : ''}`}
              >
                {workspace?.name}
              </h1>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Link
              to={`/w/${workspaceId}/members`}
              className="flex items-center gap-1 text-sm border rounded px-3 py-1.5 hover:bg-gray-100"
            >
              <Users size={16} /> Members ({members.length})
            </Link>
            {isOwner && (
              <button
                onClick={deleteWorkspace}
                className="flex items-center gap-1 text-sm border border-red-200 text-red-600 rounded px-3 py-1.5 hover:bg-red-50"
              >
                <Trash2 size={16} /> Delete
              </button>
            )}
          </div>
        </div>

        <form onSubmit={handleCreateBoard} className="flex gap-2 mb-6">
          <input
            required
            value={newBoardName}
            onChange={(e) => setNewBoardName(e.target.value)}
            placeholder="New board name"
            className="flex-1 border rounded px-3 py-2 text-sm"
          />
          <button className="flex items-center gap-1 bg-[#D6AE32] text-gray-900 px-4 py-2 rounded text-sm font-medium hover:bg-[#B6942B]">
            <Plus size={16} /> Create board
          </button>
        </form>

        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {boards.map((b) => {
            const detail = boardDetails[b.id]
            const tileIsAdmin = detail?.members.some((m) => m.user_id === user?.id && m.role === 'admin') ?? false
            return (
              <div key={b.id} className="relative group">
                <Link
                  to={`/b/${b.id}`}
                  className="h-24 rounded-lg p-4 text-white font-medium shadow-sm hover:brightness-95 transition flex items-end"
                  style={{ backgroundColor: b.background_color }}
                >
                  {b.name}
                </Link>
                {tileIsAdmin && (
                  <button
                    onClick={(e) => {
                      e.preventDefault()
                      e.stopPropagation()
                      setSettingsBoardId(b.id)
                      setSettingsAnchor({ x: e.clientX, y: e.clientY })
                    }}
                    title="Board settings"
                    className="absolute top-1.5 right-1.5 bg-white/90 rounded p-1 opacity-0 group-hover:opacity-100 hover:bg-white transition"
                  >
                    <Settings size={14} className="text-gray-700" />
                  </button>
                )}
              </div>
            )
          })}
        </div>

        {settingsBoardId && settingsAnchor && boardDetails[settingsBoardId] && (
          <BoardSettingsMenu
            board={boardDetails[settingsBoardId]}
            anchor={settingsAnchor}
            onChanged={() => {
              load()
              refreshBoardDetail(settingsBoardId)
            }}
            onDeleted={() => {
              setSettingsBoardId(null)
              setSettingsAnchor(null)
              load()
            }}
            onClose={() => {
              setSettingsBoardId(null)
              setSettingsAnchor(null)
            }}
            onShowArchived={() => setArchivedBoardId(settingsBoardId)}
          />
        )}

        {archivedBoardId && boardDetails[archivedBoardId] && (
          <ArchivedCardsModal
            board={boardDetails[archivedBoardId]}
            onClose={() => setArchivedBoardId(null)}
            onChanged={() => refreshBoardDetail(archivedBoardId)}
          />
        )}
      </main>
    </div>
  )
}
