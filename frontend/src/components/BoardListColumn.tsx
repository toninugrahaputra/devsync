import { useState, type FormEvent } from 'react'
import { useSortable, SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Plus } from 'lucide-react'
import client from '../api/client'
import SortableCard from './SortableCard'
import type { BoardDetail, ListDetail } from '../types'

interface Props {
  list: ListDetail
  board: BoardDetail
  onOpenCard: (id: number) => void
  onChanged: () => void
}

export default function BoardListColumn({ list, board, onOpenCard, onChanged }: Props) {
  const [adding, setAdding] = useState(false)
  const [title, setTitle] = useState('')
  const [renaming, setRenaming] = useState(false)
  const [name, setName] = useState(list.name)

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: `list-${list.id}`,
    data: { type: 'list', listId: list.id },
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const cardIds = list.cards.map((c) => `card-${c.id}`)

  async function handleAddCard(e: FormEvent) {
    e.preventDefault()
    if (!title.trim()) return
    await client.post(`/lists/${list.id}/cards`, { title })
    setTitle('')
    setAdding(false)
    onChanged()
  }

  async function handleRename() {
    setRenaming(false)
    if (name.trim() && name !== list.name) {
      await client.put(`/lists/${list.id}`, { name })
      onChanged()
    } else {
      setName(list.name)
    }
  }

  return (
    <div ref={setNodeRef} style={style} className="w-72 shrink-0 bg-[#ebecf0] rounded-lg p-2 flex flex-col max-h-full">
      <div {...attributes} {...listeners} className="px-1 py-1 cursor-grab active:cursor-grabbing">
        {renaming ? (
          <input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            onBlur={handleRename}
            onKeyDown={(e) => e.key === 'Enter' && handleRename()}
            className="w-full font-semibold text-sm px-1 py-0.5 rounded border"
          />
        ) : (
          <h2 onClick={() => setRenaming(true)} className="font-semibold text-sm text-gray-700 px-1 py-0.5">
            {list.name} <span className="text-gray-400 font-normal">{list.cards.length}</span>
          </h2>
        )}
      </div>

      <div className="flex-1 overflow-y-auto space-y-2 px-1 py-1 min-h-8">
        <SortableContext items={cardIds} strategy={verticalListSortingStrategy}>
          {list.cards.map((card) => (
            <SortableCard key={card.id} card={card} board={board} onOpen={() => onOpenCard(card.id)} onChanged={onChanged} />
          ))}
        </SortableContext>
      </div>

      {adding ? (
        <form onSubmit={handleAddCard} className="px-1 pt-1">
          <textarea
            autoFocus
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Enter a title for this card..."
            rows={2}
            className="w-full text-sm rounded border px-2 py-1.5 resize-none"
          />
          <div className="flex gap-2 mt-1">
            <button type="submit" className="bg-[#D6AE32] text-gray-900 text-sm px-3 py-1 rounded">
              Add card
            </button>
            <button type="button" onClick={() => setAdding(false)} className="text-sm text-gray-500 px-2">
              ✕
            </button>
          </div>
        </form>
      ) : (
        <button
          onClick={() => setAdding(true)}
          className="flex items-center gap-1 text-sm text-gray-600 px-2 py-1.5 hover:bg-gray-200 rounded"
        >
          <Plus size={14} /> Add a card
        </button>
      )}
    </div>
  )
}
