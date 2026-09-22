import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { ArrowLeft, Settings } from 'lucide-react'
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  pointerWithin,
  rectIntersection,
  useSensor,
  useSensors,
  type CollisionDetection,
  type DragEndEvent,
  type DragOverEvent,
  type DragStartEvent,
} from '@dnd-kit/core'
import { SortableContext, horizontalListSortingStrategy, arrayMove } from '@dnd-kit/sortable'
import client from '../api/client'
import ArchivedCardsModal from '../components/ArchivedCardsModal'
import BoardListColumn from '../components/BoardListColumn'
import BoardSettingsMenu from '../components/BoardSettingsMenu'
import CardChip from '../components/CardChip'
import CardModal from '../components/CardModal'
import logo from '../assets/logo.png'
import { useAuth } from '../context/AuthContext'
import type { BoardDetail, CardSummary, ListDetail } from '../types'

export default function BoardPage() {
  const { boardId } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const [board, setBoard] = useState<BoardDetail | null>(null)
  const [activeCard, setActiveCard] = useState<CardSummary | null>(null)
  const [openCardId, setOpenCardId] = useState<number | null>(searchParams.get('card') ? Number(searchParams.get('card')) : null)
  const [newListName, setNewListName] = useState('')
  const [editingTitle, setEditingTitle] = useState(false)
  const [settingsAnchor, setSettingsAnchor] = useState<{ x: number; y: number } | null>(null)
  const [showArchived, setShowArchived] = useState(false)

  function closeCard() {
    setOpenCardId(null)
    if (searchParams.has('card')) {
      searchParams.delete('card')
      setSearchParams(searchParams, { replace: true })
    }
  }

  // dnd-kit fires onDragOver repeatedly during a single drag, faster than React
  // necessarily re-renders. handleDragEnd needs the *latest* reparented state at
  // drop time, not whatever `board` the enclosing render closed over, so every
  // update goes through this ref synchronously (in lockstep with setBoard).
  const boardRef = useRef<BoardDetail | null>(null)

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }))

  // closestCorners compares corner distances across every registered droppable
  // (every card in every column, plus each column itself), which misfires across
  // columns of different heights/content. pointerWithin — "is the pointer literally
  // inside this rect" — is what a Kanban board actually needs; rectIntersection is
  // the fallback for the moment the pointer briefly clears every rect mid-drag.
  const collisionDetection: CollisionDetection = useCallback((args) => {
    const pointerCollisions = pointerWithin(args)
    if (pointerCollisions.length > 0) return pointerCollisions
    return rectIntersection(args)
  }, [])

  function updateBoard(updater: (prev: BoardDetail | null) => BoardDetail | null) {
    setBoard((prev) => {
      const next = updater(prev)
      boardRef.current = next
      return next
    })
  }

  const [loadError, setLoadError] = useState('')

  const load = useCallback(async () => {
    if (!boardId) return
    setLoadError('')
    try {
      const { data } = await client.get<BoardDetail>(`/boards/${boardId}`)
      boardRef.current = data
      setBoard(data)
    } catch {
      setLoadError("Could not load this board — it may not exist, or you may not have access to it.")
    }
  }, [boardId])

  useEffect(() => {
    load()
  }, [load])

  function findCardContainer(current: BoardDetail | null, cardId: number): ListDetail | undefined {
    return current?.lists.find((l) => l.cards.some((c) => c.id === cardId))
  }

  function handleDragStart(event: DragStartEvent) {
    const data = event.active.data.current
    if (data?.type === 'card') {
      const list = findCardContainer(boardRef.current, data.cardId as number)
      const card = list?.cards.find((c) => c.id === data.cardId)
      setActiveCard(card ?? null)
    }
  }

  function handleDragOver(event: DragOverEvent) {
    const { active, over } = event
    if (!over) return
    const activeData = active.data.current
    const overData = over.data.current
    if (activeData?.type !== 'card') return

    updateBoard((prev) => {
      if (!prev) return prev
      const sourceList = findCardContainer(prev, activeData.cardId as number)
      if (!sourceList) return prev

      let destListId: number | undefined
      if (overData?.type === 'card') destListId = overData.listId as number
      else if (overData?.type === 'list') destListId = overData.listId as number
      if (destListId === undefined || sourceList.id === destListId) return prev

      const lists = prev.lists.map((l) => ({ ...l, cards: [...l.cards] }))
      const src = lists.find((l) => l.id === sourceList.id)!
      const dst = lists.find((l) => l.id === destListId)!
      const cardIdx = src.cards.findIndex((c) => c.id === activeData.cardId)
      if (cardIdx === -1) return prev
      const [moved] = src.cards.splice(cardIdx, 1)
      const movedCopy = { ...moved, list_id: destListId }
      const overIdx = overData?.type === 'card' ? dst.cards.findIndex((c) => c.id === overData.cardId) : dst.cards.length
      dst.cards.splice(overIdx === -1 ? dst.cards.length : overIdx, 0, movedCopy)
      return { ...prev, lists }
    })
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    setActiveCard(null)
    const current = boardRef.current
    if (!over || !current) return

    const activeData = active.data.current
    const overData = over.data.current

    if (activeData?.type === 'list') {
      const listIds = current.lists.map((l) => l.id)
      const oldIndex = listIds.indexOf(activeData.listId as number)
      const newIndex = overData?.type === 'list' ? listIds.indexOf(overData.listId as number) : listIds.length - 1
      if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) return
      const reordered = arrayMove(current.lists, oldIndex, newIndex)
      updateBoard(() => ({ ...current, lists: reordered }))
      await client.put(`/lists/${activeData.listId}/move`, { index: newIndex })
      return
    }

    if (activeData?.type === 'card') {
      const destList = findCardContainer(current, activeData.cardId as number)
      if (!destList) return

      if (overData?.type === 'card' && overData.listId === destList.id) {
        const index = destList.cards.findIndex((c) => c.id === activeData.cardId)
        const overIndex = destList.cards.findIndex((c) => c.id === overData.cardId)
        if (overIndex !== -1 && overIndex !== index) {
          updateBoard((prev) => {
            if (!prev) return prev
            const lists = prev.lists.map((l) => (l.id === destList.id ? { ...l, cards: arrayMove(l.cards, index, overIndex) } : l))
            return { ...prev, lists }
          })
        }
      }

      const finalList = findCardContainer(boardRef.current, activeData.cardId as number)
      const finalIndex = finalList ? finalList.cards.findIndex((c) => c.id === activeData.cardId) : 0
      try {
        await client.put(`/cards/${activeData.cardId}/move`, {
          list_id: destList.id,
          index: finalIndex,
        })
      } catch {
        load()
      }
    }
  }

  async function renameBoard(name: string) {
    setEditingTitle(false)
    if (!board || !name.trim() || name === board.name) return
    await client.put(`/boards/${board.id}`, { name, background_color: board.background_color })
    load()
  }

  async function handleAddList(e: FormEvent) {
    e.preventDefault()
    if (!newListName.trim() || !boardId) return
    await client.post(`/boards/${boardId}/lists`, { name: newListName })
    setNewListName('')
    load()
  }

  const listIds = useMemo(() => (board ? board.lists.map((l) => `list-${l.id}`) : []), [board])
  const isAdmin = board?.members.some((m) => m.user_id === user?.id && m.role === 'admin') ?? false

  if (!board) {
    return (
      <div className="h-full bg-white">
        {loadError ? (
          <div className="p-6 text-sm text-red-600 flex items-center gap-3">
            {loadError}
            <button onClick={load} className="underline hover:no-underline">
              Retry
            </button>
          </div>
        ) : (
          <p className="text-gray-700 p-6">Loading board...</p>
        )}
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col bg-white relative">
      {/* Centered brand watermark — same on every board, so it never fights with card content */}
      <img
        src={logo}
        alt=""
        className="pointer-events-none select-none absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[30rem] opacity-10 z-0"
      />

      <div className="px-4 py-2 flex items-center gap-3 relative z-10">
        <Link
          to={`/w/${board.workspace_id}`}
          className="flex items-center gap-1 text-gray-600 hover:text-gray-900 text-sm bg-gray-100 hover:bg-gray-200 rounded px-2 py-1 transition"
        >
          <ArrowLeft size={14} /> Board
        </Link>
        <span className="w-2.5 h-2.5 rounded-full shrink-0" style={{ backgroundColor: board.background_color }} />
        {editingTitle ? (
          <input
            autoFocus
            defaultValue={board.name}
            onBlur={(e) => renameBoard(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && (e.target as HTMLInputElement).blur()}
            className="text-gray-800 font-semibold text-lg outline-none border-b border-[#D6AE32] bg-transparent"
          />
        ) : (
          <h1
            onClick={() => isAdmin && setEditingTitle(true)}
            className={`text-gray-800 font-semibold text-lg ${isAdmin ? 'cursor-text hover:bg-gray-100 rounded px-1 -mx-1' : ''}`}
          >
            {board.name}
          </h1>
        )}

        {isAdmin && (
          <button
            onClick={(e) => setSettingsAnchor({ x: e.clientX, y: e.clientY })}
            className="ml-auto flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900 bg-gray-100 hover:bg-gray-200 rounded px-2 py-1 transition"
          >
            <Settings size={14} /> Board settings
          </button>
        )}
      </div>

      {settingsAnchor && (
        <BoardSettingsMenu
          board={board}
          anchor={settingsAnchor}
          onChanged={load}
          onDeleted={() => navigate(`/w/${board.workspace_id}`)}
          onClose={() => setSettingsAnchor(null)}
          onShowArchived={() => setShowArchived(true)}
        />
      )}

      {showArchived && <ArchivedCardsModal board={board} onClose={() => setShowArchived(false)} onChanged={load} />}

      <DndContext
        sensors={sensors}
        collisionDetection={collisionDetection}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
      >
        <div className="flex-1 min-h-0 overflow-x-auto px-4 pb-4 relative z-10">
          <div className="flex gap-3 items-start h-full">
            <SortableContext items={listIds} strategy={horizontalListSortingStrategy}>
              {board.lists.map((list) => (
                <BoardListColumn key={list.id} list={list} board={board} onOpenCard={setOpenCardId} onChanged={load} />
              ))}
            </SortableContext>

            <form onSubmit={handleAddList} className="w-72 shrink-0 bg-black/5 rounded-lg p-2">
              <input
                value={newListName}
                onChange={(e) => setNewListName(e.target.value)}
                placeholder="+ Add another list"
                className="w-full bg-white border border-gray-200 rounded px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-[#D6AE32]"
              />
              {newListName && (
                <button className="mt-2 bg-[#D6AE32] text-gray-900 text-sm px-3 py-1.5 rounded">Add list</button>
              )}
            </form>
          </div>
        </div>

        <DragOverlay>{activeCard ? <CardChip card={activeCard} labels={board.labels} members={board.members} /> : null}</DragOverlay>
      </DndContext>

      {openCardId && <CardModal cardId={openCardId} board={board} onClose={closeCard} onChanged={load} />}
    </div>
  )
}
