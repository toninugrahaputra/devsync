import { useEffect, useState, type FormEvent } from 'react'
import { X, CalendarDays, Check, Tag, Users, CheckSquare, Trash2, Plus, Paperclip, Download } from 'lucide-react'
import client, { fileUrl } from '../api/client'
import Avatar from './Avatar'
import { useConfirm } from '../context/ConfirmContext'
import { formatDate } from '../utils/date'
import type { Activity, BoardDetail, CardDetail, ChecklistDetail, Comment } from '../types'

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

type FeedItem = { type: 'comment'; data: Comment } | { type: 'activity'; data: Activity }

interface Props {
  cardId: number
  board: BoardDetail
  onClose: () => void
  onChanged: () => void
}

export default function CardModal({ cardId, board, onClose, onChanged }: Props) {
  const [card, setCard] = useState<CardDetail | null>(null)
  const [description, setDescription] = useState('')
  const [editingDesc, setEditingDesc] = useState(false)
  const [commentBody, setCommentBody] = useState('')
  const [newChecklistTitle, setNewChecklistTitle] = useState('')
  const [newItemText, setNewItemText] = useState<Record<number, string>>({})
  const confirm = useConfirm()

  async function load() {
    const { data } = await client.get<CardDetail>(`/cards/${cardId}`)
    setCard(data)
    setDescription(data.description)
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cardId])

  async function refresh() {
    await load()
    onChanged()
  }

  async function saveDescription() {
    setEditingDesc(false)
    if (card && description !== card.description) {
      await client.put(`/cards/${cardId}`, { description })
      refresh()
    }
  }

  async function saveTitle(title: string) {
    if (card && title.trim() && title !== card.title) {
      await client.put(`/cards/${cardId}`, { title })
      refresh()
    }
  }

  async function toggleComplete() {
    if (!card) return
    await client.put(`/cards/${cardId}`, { is_completed: !card.is_completed })
    refresh()
  }

  async function toggleMember(userId: number) {
    if (!card) return
    const has = card.members.some((m) => m.user_id === userId)
    if (has) {
      await client.delete(`/cards/${cardId}/members/${userId}`)
    } else {
      await client.post(`/cards/${cardId}/members/${userId}`)
    }
    refresh()
  }

  async function toggleLabel(labelId: number) {
    if (!card) return
    const has = card.labels.some((l) => l.id === labelId)
    if (has) {
      await client.delete(`/cards/${cardId}/labels/${labelId}`)
    } else {
      await client.post(`/cards/${cardId}/labels/${labelId}`)
    }
    refresh()
  }

  async function setDueDate(value: string) {
    await client.put(`/cards/${cardId}`, { due_date: value || '' })
    refresh()
  }

  async function setStartDate(value: string) {
    await client.put(`/cards/${cardId}`, { start_date: value || '' })
    refresh()
  }

  async function addComment(e: FormEvent) {
    e.preventDefault()
    if (!commentBody.trim()) return
    await client.post(`/cards/${cardId}/comments`, { body: commentBody })
    setCommentBody('')
    refresh()
  }

  async function deleteComment(id: number) {
    if (!(await confirm({ title: 'Delete comment', message: 'Delete this comment?', confirmLabel: 'Delete', danger: true }))) return
    await client.delete(`/comments/${id}`)
    refresh()
  }

  async function addChecklist(e: FormEvent) {
    e.preventDefault()
    if (!newChecklistTitle.trim()) return
    await client.post(`/cards/${cardId}/checklists`, { title: newChecklistTitle })
    setNewChecklistTitle('')
    refresh()
  }

  async function deleteChecklist(id: number) {
    const ok = await confirm({
      title: 'Delete checklist',
      message: 'Delete this checklist and all its items?',
      confirmLabel: 'Delete',
      danger: true,
    })
    if (!ok) return
    await client.delete(`/checklists/${id}`)
    refresh()
  }

  async function addItem(checklistId: number) {
    const text = newItemText[checklistId]
    if (!text?.trim()) return
    await client.post(`/checklists/${checklistId}/items`, { text })
    setNewItemText((prev) => ({ ...prev, [checklistId]: '' }))
    refresh()
  }

  async function toggleItem(itemId: number, isChecked: boolean) {
    await client.put(`/checklist-items/${itemId}`, { is_checked: !isChecked })
    refresh()
  }

  async function deleteItem(itemId: number) {
    const ok = await confirm({ title: 'Delete item', message: 'Delete this checklist item?', confirmLabel: 'Delete', danger: true })
    if (!ok) return
    await client.delete(`/checklist-items/${itemId}`)
    refresh()
  }

  async function uploadAttachment(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    await client.post(`/cards/${cardId}/attachments`, formData)
    refresh()
  }

  async function deleteAttachment(id: number) {
    const ok = await confirm({ title: 'Delete attachment', message: 'Delete this attachment?', confirmLabel: 'Delete', danger: true })
    if (!ok) return
    await client.delete(`/attachments/${id}`)
    refresh()
  }

  async function deleteCard() {
    const ok = await confirm({
      title: 'Archive card',
      message: 'This card will move to the board\'s Archived items, where you can restore it or delete it for good.',
      confirmLabel: 'Archive',
      danger: true,
    })
    if (!ok) return
    await client.delete(`/cards/${cardId}`)
    onChanged()
    onClose()
  }

  function checklistProgress(cl: ChecklistDetail) {
    if (cl.items.length === 0) return 0
    return Math.round((cl.items.filter((i) => i.is_checked).length / cl.items.length) * 100)
  }

  const feed: FeedItem[] = card
    ? [
        ...card.comments.map((c): FeedItem => ({ type: 'comment', data: c })),
        ...card.activities.map((a): FeedItem => ({ type: 'activity', data: a })),
      ].sort((a, b) => new Date(b.data.created_at).getTime() - new Date(a.data.created_at).getTime())
    : []

  return (
    <div className="fixed inset-0 bg-black/50 flex items-start justify-center overflow-y-auto py-10 z-50" onClick={onClose}>
      <div className="bg-white rounded-lg w-full max-w-2xl shadow-xl overflow-hidden" onClick={(e) => e.stopPropagation()}>
        {!card ? (
          <div className="p-6 text-sm text-gray-500">Loading...</div>
        ) : (
          <div>
            {card.cover_color && <div className="h-24" style={{ backgroundColor: card.cover_color }} />}
            <div className="p-5">
            <div className="flex items-start justify-between gap-2">
              <button
                onClick={toggleComplete}
                title={card.is_completed ? 'Mark incomplete' : 'Mark complete'}
                className={`shrink-0 mt-1 w-5 h-5 rounded-full border-2 flex items-center justify-center transition ${
                  card.is_completed
                    ? 'bg-green-600 border-green-600 text-white'
                    : 'border-gray-300 text-transparent hover:border-green-500 hover:text-green-500'
                }`}
              >
                <Check size={12} strokeWidth={3} />
              </button>
              <input
                defaultValue={card.title}
                onBlur={(e) => saveTitle(e.target.value)}
                className={`text-lg font-semibold flex-1 outline-none rounded px-1 py-0.5 hover:bg-gray-100 focus:bg-gray-100 ${
                  card.is_completed ? 'text-gray-400 line-through' : ''
                }`}
              />
              <button onClick={onClose} className="text-gray-400 hover:text-gray-700 p-1">
                <X size={20} />
              </button>
            </div>

            <div className="flex gap-2 mt-4 flex-wrap">
              <details className="relative">
                <summary className="list-none flex items-center gap-1 text-xs border rounded px-2 py-1.5 cursor-pointer hover:bg-gray-100">
                  <Users size={14} /> Members
                </summary>
                <div className="absolute z-10 bg-white border rounded shadow mt-1 w-48 max-h-52 overflow-auto p-1">
                  {board.members.map((m) => (
                    <label
                      key={m.user_id}
                      className="flex items-center gap-2 px-2 py-1.5 text-sm hover:bg-gray-100 rounded cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={card.members.some((cm) => cm.user_id === m.user_id)}
                        onChange={() => toggleMember(m.user_id)}
                      />
                      {m.name}
                    </label>
                  ))}
                </div>
              </details>

              <details className="relative">
                <summary className="list-none flex items-center gap-1 text-xs border rounded px-2 py-1.5 cursor-pointer hover:bg-gray-100">
                  <Tag size={14} /> Labels
                </summary>
                <div className="absolute z-10 bg-white border rounded shadow mt-1 w-48 max-h-52 overflow-auto p-1">
                  {board.labels.map((l) => (
                    <label
                      key={l.id}
                      className="flex items-center gap-2 px-2 py-1.5 text-sm hover:bg-gray-100 rounded cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={card.labels.some((cl) => cl.id === l.id)}
                        onChange={() => toggleLabel(l.id)}
                      />
                      <span className="w-4 h-4 rounded" style={{ backgroundColor: l.color }} />
                      {l.name || l.color}
                    </label>
                  ))}
                </div>
              </details>

              <details className="relative">
                <summary className="list-none flex items-center gap-1 text-xs border rounded px-2 py-1.5 cursor-pointer hover:bg-gray-100">
                  <CalendarDays size={14} /> Dates
                </summary>
                <div className="absolute z-10 bg-white border rounded shadow mt-1 w-56 p-3 space-y-2">
                  <div>
                    <label className="text-xs text-gray-500">Start date</label>
                    <input
                      type="date"
                      defaultValue={card.start_date?.slice(0, 10) || ''}
                      onChange={(e) => setStartDate(e.target.value)}
                      className="w-full border rounded px-2 py-1 text-sm mt-0.5"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-gray-500">Due date</label>
                    <input
                      type="date"
                      defaultValue={card.due_date?.slice(0, 10) || ''}
                      onChange={(e) => setDueDate(e.target.value)}
                      className="w-full border rounded px-2 py-1 text-sm mt-0.5"
                    />
                  </div>
                </div>
              </details>

              <button
                onClick={deleteCard}
                className="flex items-center gap-1 text-xs border rounded px-2 py-1.5 hover:bg-red-50 text-red-600 ml-auto"
              >
                <Trash2 size={14} /> Archive
              </button>
            </div>

            {(card.labels.length > 0 || card.members.length > 0 || card.due_date || card.start_date) && (
              <div className="flex flex-wrap gap-4 mt-3 text-xs">
                {card.labels.length > 0 && (
                  <div>
                    <p className="text-gray-500 mb-1">Labels</p>
                    <div className="flex gap-1 flex-wrap">
                      {card.labels.map((l) => (
                        <span key={l.id} className="text-gray-900 rounded px-2 py-1" style={{ backgroundColor: l.color }}>
                          {l.name || l.color}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                {card.members.length > 0 && (
                  <div>
                    <p className="text-gray-500 mb-1">Members</p>
                    <div className="flex gap-1">
                      {card.members.map((m) => (
                        <Avatar key={m.user_id} name={m.name} avatarPath={m.avatar_path} size={24} />
                      ))}
                    </div>
                  </div>
                )}
                {(card.due_date || card.start_date) && (
                  <div>
                    <p className="text-gray-500 mb-1">Dates</p>
                    <span className="border rounded px-2 py-1">
                      {card.start_date ? formatDate(card.start_date) : '...'} →{' '}
                      {card.due_date ? formatDate(card.due_date) : '...'}
                    </span>
                  </div>
                )}
              </div>
            )}

            <div className="mt-5">
              <h3 className="text-sm font-semibold text-gray-700 mb-1">Description</h3>
              {editingDesc ? (
                <div>
                  <textarea
                    autoFocus
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={3}
                    className="w-full border rounded px-2 py-1.5 text-sm"
                  />
                  <div className="flex gap-2 mt-1">
                    <button onClick={saveDescription} className="bg-[#D6AE32] text-gray-900 text-xs px-3 py-1.5 rounded">
                      Save
                    </button>
                    <button
                      onClick={() => {
                        setEditingDesc(false)
                        setDescription(card.description)
                      }}
                      className="text-xs text-gray-500 px-2"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <div
                  onClick={() => setEditingDesc(true)}
                  className="text-sm text-gray-700 bg-gray-100 rounded px-2 py-2 min-h-10 cursor-text hover:bg-gray-200"
                >
                  {card.description || <span className="text-gray-400">Add a more detailed description...</span>}
                </div>
              )}
            </div>

            <div className="mt-5">
              <h3 className="text-sm font-semibold text-gray-700 mb-1 flex items-center gap-1">
                <CheckSquare size={16} /> Checklists
              </h3>
              {card.checklists.map((cl) => (
                <div key={cl.id} className="mb-4">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-gray-700">{cl.title}</p>
                    <button onClick={() => deleteChecklist(cl.id)} className="text-gray-400 hover:text-red-500">
                      <Trash2 size={14} />
                    </button>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-1.5 my-1.5">
                    <div className="bg-green-500 h-1.5 rounded-full" style={{ width: `${checklistProgress(cl)}%` }} />
                  </div>
                  <div className="space-y-1">
                    {cl.items.map((item) => (
                      <div key={item.id} className="flex items-center gap-2 group">
                        <input type="checkbox" checked={item.is_checked} onChange={() => toggleItem(item.id, item.is_checked)} />
                        <span className={`text-sm flex-1 ${item.is_checked ? 'line-through text-gray-400' : 'text-gray-700'}`}>
                          {item.text}
                        </span>
                        <button
                          onClick={() => deleteItem(item.id)}
                          className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500"
                        >
                          <Trash2 size={12} />
                        </button>
                      </div>
                    ))}
                  </div>
                  <div className="flex gap-1 mt-1.5">
                    <input
                      value={newItemText[cl.id] || ''}
                      onChange={(e) => setNewItemText((prev) => ({ ...prev, [cl.id]: e.target.value }))}
                      onKeyDown={(e) => e.key === 'Enter' && addItem(cl.id)}
                      placeholder="Add an item"
                      className="flex-1 border rounded px-2 py-1 text-sm"
                    />
                    <button onClick={() => addItem(cl.id)} className="text-xs border rounded px-2 hover:bg-gray-100">
                      Add
                    </button>
                  </div>
                </div>
              ))}
              <form onSubmit={addChecklist} className="flex gap-1">
                <input
                  required
                  value={newChecklistTitle}
                  onChange={(e) => setNewChecklistTitle(e.target.value)}
                  placeholder="Add a checklist"
                  className="flex-1 border rounded px-2 py-1.5 text-sm"
                />
                <button className="flex items-center gap-1 text-xs border rounded px-2 hover:bg-gray-100">
                  <Plus size={14} /> Add
                </button>
              </form>
            </div>

            <div className="mt-5">
              <h3 className="text-sm font-semibold text-gray-700 mb-1 flex items-center gap-1">
                <Paperclip size={16} /> Attachments
              </h3>
              <div className="space-y-1.5 mb-2">
                {card.attachments.map((a) => (
                  <div key={a.id} className="flex items-center justify-between bg-gray-100 rounded px-2.5 py-1.5 group">
                    <a
                      href={fileUrl(a.file_path)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 text-sm text-gray-800 hover:text-[#B6942B] min-w-0"
                    >
                      <Download size={14} className="shrink-0" />
                      <span className="truncate">{a.file_name}</span>
                      <span className="text-xs text-gray-400 shrink-0">{formatFileSize(a.file_size)}</span>
                    </a>
                    <button
                      onClick={() => deleteAttachment(a.id)}
                      className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500 shrink-0"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
              </div>
              <label className="inline-flex items-center gap-1.5 text-xs border rounded px-2 py-1.5 hover:bg-gray-50 cursor-pointer">
                <Plus size={14} /> Add attachment
                <input
                  type="file"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0]
                    if (file) uploadAttachment(file)
                    e.target.value = ''
                  }}
                />
              </label>
            </div>

            <div className="mt-5">
              <h3 className="text-sm font-semibold text-gray-700 mb-2">Comments and activity</h3>
              <form onSubmit={addComment} className="flex gap-2 mb-3">
                <input
                  required
                  value={commentBody}
                  onChange={(e) => setCommentBody(e.target.value)}
                  placeholder="Write a comment..."
                  className="flex-1 border rounded px-2 py-1.5 text-sm"
                />
                <button className="text-xs bg-[#D6AE32] text-gray-900 px-3 py-1.5 rounded">Send</button>
              </form>
              <div className="space-y-2">
                {feed.map((item) =>
                  item.type === 'comment' ? (
                    <div key={`comment-${item.data.id}`} className="bg-gray-100 rounded p-2 group">
                      <div className="flex items-center justify-between">
                        <p className="text-xs font-medium text-gray-700">{item.data.user_name}</p>
                        <div className="flex items-center gap-2">
                          <span className="text-[10px] text-gray-400">{new Date(item.data.created_at).toLocaleString()}</span>
                          <button
                            onClick={() => deleteComment(item.data.id)}
                            className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500"
                          >
                            <Trash2 size={12} />
                          </button>
                        </div>
                      </div>
                      <p className="text-sm text-gray-800 mt-0.5">{item.data.body}</p>
                    </div>
                  ) : (
                    <div key={`activity-${item.data.id}`} className="text-xs text-gray-500 px-1 py-0.5">
                      <span className="font-medium text-gray-600">{item.data.user_name}</span> {item.data.message}
                      <span className="text-gray-400"> · {new Date(item.data.created_at).toLocaleString()}</span>
                    </div>
                  )
                )}
              </div>
            </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
