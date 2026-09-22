import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Archive, ArrowLeft, Palette, Trash2, Users } from 'lucide-react'
import client from '../api/client'
import { useConfirm } from '../context/ConfirmContext'
import type { BoardDetail, WorkspaceMember } from '../types'

const BOARD_COLORS = [
  '#0079BF',
  '#D29034',
  '#519839',
  '#B04632',
  '#89609E',
  '#CD5A91',
  '#4BBF6B',
  '#00AECC',
  '#B9871A',
]

type View = 'menu' | 'color' | 'members'

interface Props {
  board: BoardDetail
  anchor: { x: number; y: number }
  onChanged: () => void
  onDeleted: () => void
  onClose: () => void
  onShowArchived?: () => void
}

export default function BoardSettingsMenu({ board, anchor, onChanged, onDeleted, onClose, onShowArchived }: Props) {
  const [view, setView] = useState<View>('menu')
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMember[]>([])
  const confirm = useConfirm()

  useEffect(() => {
    if (view !== 'members') return
    client.get<WorkspaceMember[]>(`/workspaces/${board.workspace_id}/members`).then(({ data }) => setWorkspaceMembers(data))
  }, [view, board.workspace_id])

  async function setColor(color: string) {
    await client.put(`/boards/${board.id}`, { name: board.name, background_color: color })
    onChanged()
    onClose()
  }

  async function toggleMember(userId: number) {
    const isMember = board.members.some((m) => m.user_id === userId)
    if (isMember) {
      await client.delete(`/boards/${board.id}/members/${userId}`)
    } else {
      await client.post(`/boards/${board.id}/members`, { user_id: userId })
    }
    onChanged()
  }

  async function deleteBoard() {
    const ok = await confirm({
      title: 'Delete board',
      message: `Delete "${board.name}"? This removes every list and card on it. This can't be undone.`,
      confirmLabel: 'Delete',
      danger: true,
    })
    if (!ok) return
    await client.delete(`/boards/${board.id}`)
    onDeleted()
  }

  const panelStyle = {
    position: 'fixed' as const,
    top: Math.min(anchor.y, window.innerHeight - 360),
    left: Math.min(anchor.x, window.innerWidth - 260),
  }

  function Header({ title }: { title: string }) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 border-b">
        <button onClick={() => setView('menu')} className="text-gray-500 hover:text-gray-800">
          <ArrowLeft size={16} />
        </button>
        <p className="text-sm font-medium text-gray-700">{title}</p>
      </div>
    )
  }

  return createPortal(
    <div className="fixed inset-0 z-50" onClick={(e) => { e.stopPropagation(); onClose(); }}>
      <div
        style={panelStyle}
        onClick={(e) => e.stopPropagation()}
        className="w-64 bg-white rounded-lg shadow-xl border overflow-hidden"
      >
        {view === 'menu' && (
          <div className="py-1">
            <MenuItem icon={<Palette size={15} />} label="Change color" onClick={() => setView('color')} />
            <MenuItem icon={<Users size={15} />} label="Board members" onClick={() => setView('members')} />
            {onShowArchived && (
              <MenuItem
                icon={<Archive size={15} />}
                label="Archived items"
                onClick={() => {
                  onShowArchived()
                  onClose()
                }}
              />
            )}
            <div className="border-t my-1" />
            <MenuItem icon={<Trash2 size={15} />} label="Delete board" onClick={deleteBoard} danger />
          </div>
        )}

        {view === 'color' && (
          <div>
            <Header title="Board color" />
            <div className="p-3">
              <div className="grid grid-cols-5 gap-1.5">
                {BOARD_COLORS.map((color) => (
                  <button
                    key={color}
                    onClick={() => setColor(color)}
                    className="h-8 rounded border border-black/10 hover:opacity-80"
                    style={{ backgroundColor: color }}
                  />
                ))}
              </div>
            </div>
          </div>
        )}

        {view === 'members' && (
          <div>
            <Header title="Board members" />
            <div className="p-2 max-h-64 overflow-auto">
              {workspaceMembers.map((m) => (
                <label
                  key={m.user_id}
                  className="flex items-center gap-2 px-2 py-1.5 text-sm hover:bg-gray-100 rounded cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={board.members.some((bm) => bm.user_id === m.user_id)}
                    onChange={() => toggleMember(m.user_id)}
                  />
                  {m.name}
                </label>
              ))}
              {workspaceMembers.length === 0 && <p className="text-xs text-gray-400 px-2 py-1.5">No workspace members yet.</p>}
            </div>
          </div>
        )}
      </div>
    </div>,
    document.body
  )
}

function MenuItem({
  icon,
  label,
  onClick,
  danger,
}: {
  icon: React.ReactNode
  label: string
  onClick: () => void
  danger?: boolean
}) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2.5 w-full text-left px-3 py-2 text-sm hover:bg-gray-100 ${
        danger ? 'text-red-600' : 'text-gray-700'
      }`}
    >
      {icon}
      {label}
    </button>
  )
}
