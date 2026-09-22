import { useState, useRef } from 'react'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Check, Pencil } from 'lucide-react'
import client from '../api/client'
import CardChip from './CardChip'
import CardQuickMenu from './CardQuickMenu'
import type { BoardDetail, CardSummary } from '../types'

interface Props {
  card: CardSummary
  board: BoardDetail
  onOpen: () => void
  onChanged: () => void
}

export default function SortableCard({ card, board, onOpen, onChanged }: Props) {
  const [menuAnchor, setMenuAnchor] = useState<{ x: number; y: number } | null>(null)
  const cardRef = useRef<HTMLDivElement>(null)

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: `card-${card.id}`,
    data: { type: 'card', cardId: card.id, listId: card.list_id },
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  }

  function openMenu(e: React.MouseEvent) {
    e.stopPropagation()
    const rect = cardRef.current?.getBoundingClientRect()
    setMenuAnchor({ x: rect ? rect.right - 256 : e.clientX, y: rect ? rect.top : e.clientY })
  }

  async function toggleComplete(e: React.MouseEvent) {
    e.stopPropagation()
    await client.put(`/cards/${card.id}`, { is_completed: !card.is_completed })
    onChanged()
  }

  return (
    <div
      ref={(el) => {
        setNodeRef(el)
        cardRef.current = el
      }}
      style={style}
      {...attributes}
      {...listeners}
      onClick={onOpen}
      className="relative group"
    >
      <CardChip card={card} labels={board.labels} members={board.members} />
      <button
        onPointerDown={(e) => e.stopPropagation()}
        onClick={toggleComplete}
        title={card.is_completed ? 'Mark incomplete' : 'Mark complete'}
        className={`absolute top-1.5 left-1.5 rounded-full p-1 transition ${
          card.is_completed
            ? 'bg-green-600 text-white opacity-100'
            : 'bg-white/90 border text-transparent group-hover:text-gray-500 opacity-0 group-hover:opacity-100 hover:bg-gray-100'
        }`}
      >
        <Check size={12} />
      </button>
      <button
        onPointerDown={(e) => e.stopPropagation()}
        onClick={openMenu}
        className="absolute top-1.5 right-1.5 bg-white/90 border rounded p-1 opacity-0 group-hover:opacity-100 hover:bg-gray-100 transition"
        title="Edit card"
      >
        <Pencil size={12} />
      </button>

      {menuAnchor && (
        <CardQuickMenu
          card={card}
          board={board}
          anchor={menuAnchor}
          onOpenCard={() => {
            setMenuAnchor(null)
            onOpen()
          }}
          onChanged={onChanged}
          onClose={() => setMenuAnchor(null)}
        />
      )}
    </div>
  )
}
