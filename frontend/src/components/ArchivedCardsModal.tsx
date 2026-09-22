import { useEffect, useState } from 'react'
import { Archive, RotateCcw, Trash2, X } from 'lucide-react'
import client from '../api/client'
import { useConfirm } from '../context/ConfirmContext'
import type { ArchivedCard, BoardDetail } from '../types'

interface Props {
  board: BoardDetail
  onClose: () => void
  onChanged: () => void
}

export default function ArchivedCardsModal({ board, onClose, onChanged }: Props) {
  const [cards, setCards] = useState<ArchivedCard[] | null>(null)
  const confirm = useConfirm()

  async function load() {
    const { data } = await client.get<ArchivedCard[]>(`/boards/${board.id}/archived-cards`)
    setCards(data)
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [board.id])

  function listName(listId: number) {
    return board.lists.find((l) => l.id === listId)?.name ?? 'a list that no longer exists'
  }

  async function restore(card: ArchivedCard) {
    await client.put(`/cards/${card.id}/restore`)
    setCards((prev) => prev?.filter((c) => c.id !== card.id) ?? null)
    onChanged()
  }

  async function deleteForever(card: ArchivedCard) {
    const ok = await confirm({
      title: 'Delete card forever',
      message: `Permanently delete "${card.title}"? This can't be undone.`,
      confirmLabel: 'Delete forever',
      danger: true,
    })
    if (!ok) return
    await client.delete(`/cards/${card.id}/permanent`)
    setCards((prev) => prev?.filter((c) => c.id !== card.id) ?? null)
    onChanged()
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-start justify-center overflow-y-auto py-10 z-50" onClick={onClose}>
      <div className="bg-white rounded-lg w-full max-w-lg shadow-xl overflow-hidden" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-3 border-b">
          <h2 className="text-sm font-semibold text-gray-800 flex items-center gap-1.5">
            <Archive size={16} /> Archived items
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-700 p-1">
            <X size={18} />
          </button>
        </div>

        <div className="p-4 max-h-[70vh] overflow-y-auto">
          {cards === null && <p className="text-sm text-gray-500">Loading...</p>}
          {cards !== null && cards.length === 0 && (
            <p className="text-sm text-gray-400">No archived cards. Cards you archive from this board will show up here.</p>
          )}
          <div className="space-y-2">
            {cards?.map((card) => (
              <div key={card.id} className="flex items-center justify-between gap-3 border rounded px-3 py-2">
                <div className="min-w-0">
                  <p className="text-sm text-gray-800 truncate">{card.title}</p>
                  <p className="text-xs text-gray-400">from {listName(card.list_id)}</p>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <button
                    onClick={() => restore(card)}
                    title="Send back to board"
                    className="flex items-center gap-1 text-xs border rounded px-2 py-1 hover:bg-gray-50 text-gray-700"
                  >
                    <RotateCcw size={13} /> Restore
                  </button>
                  <button
                    onClick={() => deleteForever(card)}
                    title="Delete forever"
                    className="flex items-center gap-1 text-xs border rounded px-2 py-1 hover:bg-red-50 text-red-600"
                  >
                    <Trash2 size={13} /> Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
