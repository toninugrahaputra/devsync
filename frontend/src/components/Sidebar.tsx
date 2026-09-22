import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ChevronDown, ChevronRight } from 'lucide-react'
import client from '../api/client'
import type { Board, Workspace } from '../types'

export default function Sidebar() {
  const { workspaceId, boardId } = useParams()
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [boardsByWorkspace, setBoardsByWorkspace] = useState<Record<number, Board[]>>({})
  const [expanded, setExpanded] = useState<Set<number>>(new Set())

  useEffect(() => {
    client.get<Workspace[]>('/workspaces').then(({ data }) => setWorkspaces(data))
  }, [])

  async function loadBoards(id: number) {
    const { data } = await client.get<Board[]>(`/workspaces/${id}/boards`)
    setBoardsByWorkspace((prev) => ({ ...prev, [id]: data }))
  }

  // Keep the active workspace (from the URL) expanded and its boards fresh,
  // so creating/renaming a board and navigating back here reflects it.
  useEffect(() => {
    const activeId = workspaceId ? Number(workspaceId) : null
    if (!activeId) return
    setExpanded((prev) => new Set(prev).add(activeId))
    loadBoards(activeId)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workspaceId])

  function toggleWorkspace(id: number) {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
        if (!boardsByWorkspace[id]) loadBoards(id)
      }
      return next
    })
  }

  return (
    <aside className="w-60 shrink-0 bg-white border-r border-gray-200 overflow-y-auto hidden sm:block">
      <nav className="p-2">
        <p className="text-xs font-semibold text-gray-400 uppercase px-2 py-2">Workspaces</p>
        {workspaces.map((w) => {
          const isExpanded = expanded.has(w.id)
          const isActiveWorkspace = Number(workspaceId) === w.id

          return (
            <div key={w.id} className="mb-0.5">
              <div className="flex items-center">
                <button onClick={() => toggleWorkspace(w.id)} className="p-1 text-gray-400 hover:text-gray-700">
                  {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                </button>
                <Link
                  to={`/w/${w.id}`}
                  className={`flex-1 text-sm px-1.5 py-1.5 rounded truncate ${
                    isActiveWorkspace ? 'bg-[#D6AE32]/20 text-gray-900 font-medium' : 'text-gray-700 hover:bg-gray-100'
                  }`}
                >
                  {w.name}
                </Link>
              </div>

              {isExpanded && (
                <div className="ml-6 mt-0.5 mb-1 space-y-0.5">
                  {(boardsByWorkspace[w.id] || []).map((b) => (
                    <Link
                      key={b.id}
                      to={`/b/${b.id}`}
                      className={`flex items-center gap-1.5 text-sm px-1.5 py-1 rounded truncate ${
                        Number(boardId) === b.id ? 'bg-[#D6AE32]/20 text-gray-900 font-medium' : 'text-gray-600 hover:bg-gray-100'
                      }`}
                    >
                      <span className="w-2.5 h-2.5 rounded-sm shrink-0" style={{ backgroundColor: b.background_color }} />
                      <span className="truncate">{b.name}</span>
                    </Link>
                  ))}
                  {boardsByWorkspace[w.id]?.length === 0 && <p className="text-xs text-gray-400 px-1.5 py-1">No boards yet</p>}
                </div>
              )}
            </div>
          )
        })}
      </nav>
    </aside>
  )
}
