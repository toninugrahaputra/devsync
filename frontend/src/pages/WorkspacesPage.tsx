import { useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import client from '../api/client'
import type { Workspace } from '../types'

export default function WorkspacesPage() {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  async function load() {
    setLoading(true)
    setLoadError('')
    try {
      const { data } = await client.get<Workspace[]>('/workspaces')
      setWorkspaces(data)
    } catch {
      setLoadError('Could not load your workspaces. Check your connection and try again.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    if (!newName.trim()) return
    setCreating(true)
    try {
      await client.post('/workspaces', { name: newName })
      setNewName('')
      await load()
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-gray-50">
      <main className="max-w-4xl mx-auto p-6">
        <h1 className="text-xl font-semibold mb-4">Your workspaces</h1>

        <form onSubmit={handleCreate} className="flex gap-2 mb-6">
          <input
            required
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="New workspace name"
            className="flex-1 border rounded px-3 py-2 text-sm"
          />
          <button
            disabled={creating}
            className="flex items-center gap-1 bg-[#D6AE32] text-gray-900 px-4 py-2 rounded text-sm font-medium hover:bg-[#B6942B] disabled:opacity-50"
          >
            <Plus size={16} /> Create
          </button>
        </form>

        {loading ? (
          <p className="text-gray-500 text-sm">Loading...</p>
        ) : loadError ? (
          <div className="text-sm text-red-600 flex items-center gap-3">
            {loadError}
            <button onClick={load} className="underline hover:no-underline">
              Retry
            </button>
          </div>
        ) : workspaces.length === 0 ? (
          <p className="text-gray-500 text-sm">No workspaces yet. Create one above.</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
            {workspaces.map((w) => (
              <Link
                key={w.id}
                to={`/w/${w.id}`}
                className="block bg-white border rounded-lg p-4 shadow-sm hover:shadow-md transition"
              >
                <p className="font-medium text-gray-800">{w.name}</p>
                <p className="text-xs text-gray-400 mt-1 uppercase">{w.role}</p>
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
