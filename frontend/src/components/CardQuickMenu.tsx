import { useState } from 'react'
import { createPortal } from 'react-dom'
import {
  ArrowLeft,
  Archive,
  CalendarDays,
  Copy,
  ExternalLink,
  Image,
  Link as LinkIcon,
  Move as MoveIcon,
  Tag,
  Users,
} from 'lucide-react'
import client from '../api/client'
import { useConfirm } from '../context/ConfirmContext'
import type { BoardDetail, CardSummary } from '../types'

const COVER_COLORS = [
  '#4bce97',
  '#f5cd47',
  '#fea362',
  '#f87462',
  '#9f8fef',
  '#579dff',
  '#6cc3e0',
  '#94c748',
  '#e774bb',
  '#8590a2',
]

type View = 'menu' | 'labels' | 'members' | 'dates' | 'cover' | 'move'

interface Props {
  card: CardSummary
  board: BoardDetail
  anchor: { x: number; y: number }
  onOpenCard: () => void
  onChanged: () => void
  onClose: () => void
}

export default function CardQuickMenu({ card, board, anchor, onOpenCard, onChanged, onClose }: Props) {
  const [view, setView] = useState<View>('menu')
  const confirm = useConfirm()

  async function refresh() {
    onChanged()
  }

  async function toggleLabel(labelId: number) {
    const has = card.label_ids.includes(labelId)
    if (has) {
      await client.delete(`/cards/${card.id}/labels/${labelId}`)
    } else {
      await client.post(`/cards/${card.id}/labels/${labelId}`)
    }
    refresh()
  }

  async function toggleMember(userId: number) {
    const has = card.member_ids.includes(userId)
    if (has) {
      await client.delete(`/cards/${card.id}/members/${userId}`)
    } else {
      await client.post(`/cards/${card.id}/members/${userId}`)
    }
    refresh()
  }

  async function setDates(field: 'due_date' | 'start_date', value: string) {
    await client.put(`/cards/${card.id}`, { [field]: value || '' })
    refresh()
  }

  async function setCover(color: string | null) {
    await client.put(`/cards/${card.id}`, { cover_color: color || '' })
    refresh()
    onClose()
  }

  async function moveTo(listId: number) {
    const targetList = board.lists.find((l) => l.id === listId)
    const index = targetList ? targetList.cards.length : 0
    await client.put(`/cards/${card.id}/move`, { list_id: listId, index })
    refresh()
    onClose()
  }

  async function duplicateCard() {
    const { data: created } = await client.post(`/lists/${card.list_id}/cards`, { title: `${card.title} (copy)` })
    if (card.description) {
      await client.put(`/cards/${created.id}`, { description: card.description })
    }
    await Promise.all([
      ...card.label_ids.map((labelId) => client.post(`/cards/${created.id}/labels/${labelId}`)),
      ...card.member_ids.map((userId) => client.post(`/cards/${created.id}/members/${userId}`)),
    ])
    refresh()
    onClose()
  }

  async function copyLink() {
    const url = `${window.location.origin}/b/${card.board_id}?card=${card.id}`
    try {
      await navigator.clipboard.writeText(url)
    } catch {
      // clipboard permission denied; nothing more we can do here
    }
    onClose()
  }

  async function archiveCard() {
    const ok = await confirm({
      title: 'Archive card',
      message: 'This card will move to the board\'s Archived items, where you can restore it or delete it for good.',
      confirmLabel: 'Archive',
      danger: true,
    })
    if (!ok) return
    await client.delete(`/cards/${card.id}`)
    refresh()
    onClose()
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
    // React bubbles portal events through the *component* tree, not the DOM tree —
    // this backdrop is still a React child of the card wrapper (whose onClick opens
    // the full modal), so every click here must stop propagation or it'll open the
    // card behind the menu it's meant to be closing.
    <div className="fixed inset-0 z-50" onClick={(e) => { e.stopPropagation(); onClose(); }}>
      <div
        style={panelStyle}
        onClick={(e) => e.stopPropagation()}
        className="w-64 bg-white rounded-lg shadow-xl border overflow-hidden"
      >
        {view === 'menu' && (
          <div className="py-1">
            <MenuItem icon={<ExternalLink size={15} />} label="Open card" onClick={onOpenCard} />
            <MenuItem icon={<Tag size={15} />} label="Edit labels" onClick={() => setView('labels')} />
            <MenuItem icon={<Users size={15} />} label="Change members" onClick={() => setView('members')} />
            <MenuItem icon={<Image size={15} />} label="Change cover" onClick={() => setView('cover')} />
            <MenuItem icon={<CalendarDays size={15} />} label="Edit dates" onClick={() => setView('dates')} />
            <MenuItem icon={<LinkIcon size={15} />} label="Copy link" onClick={copyLink} />
            <MenuItem icon={<MoveIcon size={15} />} label="Move" onClick={() => setView('move')} />
            <MenuItem icon={<Copy size={15} />} label="Copy card" onClick={duplicateCard} />
            <div className="border-t my-1" />
            <MenuItem icon={<Archive size={15} />} label="Archive" onClick={archiveCard} danger />
          </div>
        )}

        {view === 'labels' && (
          <div>
            <Header title="Labels" />
            <div className="p-2 max-h-64 overflow-auto">
              {board.labels.map((l) => (
                <label key={l.id} className="flex items-center gap-2 px-2 py-1.5 text-sm hover:bg-gray-100 rounded cursor-pointer">
                  <input type="checkbox" checked={card.label_ids.includes(l.id)} onChange={() => toggleLabel(l.id)} />
                  <span className="w-4 h-4 rounded" style={{ backgroundColor: l.color }} />
                  {l.name || l.color}
                </label>
              ))}
              {board.labels.length === 0 && <p className="text-xs text-gray-400 px-2 py-1.5">No labels on this board yet.</p>}
            </div>
          </div>
        )}

        {view === 'members' && (
          <div>
            <Header title="Members" />
            <div className="p-2 max-h-64 overflow-auto">
              {board.members.map((m) => (
                <label
                  key={m.user_id}
                  className="flex items-center gap-2 px-2 py-1.5 text-sm hover:bg-gray-100 rounded cursor-pointer"
                >
                  <input type="checkbox" checked={card.member_ids.includes(m.user_id)} onChange={() => toggleMember(m.user_id)} />
                  {m.name}
                </label>
              ))}
            </div>
          </div>
        )}

        {view === 'dates' && (
          <div>
            <Header title="Dates" />
            <div className="p-3 space-y-2">
              <div>
                <label className="text-xs text-gray-500">Start date</label>
                <input
                  type="date"
                  defaultValue={card.start_date?.slice(0, 10) || ''}
                  onChange={(e) => setDates('start_date', e.target.value)}
                  className="w-full border rounded px-2 py-1 text-sm mt-0.5"
                />
              </div>
              <div>
                <label className="text-xs text-gray-500">Due date</label>
                <input
                  type="date"
                  defaultValue={card.due_date?.slice(0, 10) || ''}
                  onChange={(e) => setDates('due_date', e.target.value)}
                  className="w-full border rounded px-2 py-1 text-sm mt-0.5"
                />
              </div>
            </div>
          </div>
        )}

        {view === 'cover' && (
          <div>
            <Header title="Cover" />
            <div className="p-3">
              <div className="grid grid-cols-5 gap-1.5">
                {COVER_COLORS.map((color) => (
                  <button
                    key={color}
                    onClick={() => setCover(color)}
                    className="h-8 rounded border border-black/10 hover:opacity-80"
                    style={{ backgroundColor: color }}
                  />
                ))}
              </div>
              {card.cover_color && (
                <button onClick={() => setCover(null)} className="mt-2 text-xs text-gray-500 hover:underline">
                  Remove cover
                </button>
              )}
            </div>
          </div>
        )}

        {view === 'move' && (
          <div>
            <Header title="Move card" />
            <div className="p-2 max-h-64 overflow-auto">
              {board.lists
                .filter((l) => l.id !== card.list_id)
                .map((l) => (
                  <button
                    key={l.id}
                    onClick={() => moveTo(l.id)}
                    className="block w-full text-left px-2 py-1.5 text-sm hover:bg-gray-100 rounded"
                  >
                    {l.name}
                  </button>
                ))}
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
