import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Camera } from 'lucide-react'
import client from '../api/client'
import Avatar from '../components/Avatar'
import { useAuth } from '../context/AuthContext'

export default function ProfilePage() {
  const { user, updateUser } = useAuth()
  const [name, setName] = useState(user?.name || '')
  const [username, setUsername] = useState(user?.username || '')
  const [saving, setSaving] = useState(false)
  const [uploadingAvatar, setUploadingAvatar] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  if (!user) return null

  async function handleSave(e: FormEvent) {
    e.preventDefault()
    setError('')
    setSuccess('')
    setSaving(true)
    try {
      const { data } = await client.put('/users/me', { name, username: username.trim() })
      updateUser(data)
      setSuccess('Profile updated.')
    } catch (err) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setError(msg || 'Could not update profile')
    } finally {
      setSaving(false)
    }
  }

  async function handleAvatarChange(file: File) {
    setError('')
    setUploadingAvatar(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await client.post('/users/me/avatar', formData)
      updateUser(data)
    } catch (err) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setError(msg || 'Could not upload photo')
    } finally {
      setUploadingAvatar(false)
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-gray-50">
      <main className="max-w-lg mx-auto p-6">
        <Link to="/" className="flex items-center gap-1 text-sm text-[#B6942B] hover:underline w-fit">
          <ArrowLeft size={14} /> Back
        </Link>
        <h1 className="text-xl font-semibold mt-2 mb-6">Your profile</h1>

        <div className="bg-white border rounded-lg p-6">
          <div className="flex items-center gap-4 mb-6">
            <label className="relative cursor-pointer group">
              <Avatar name={user.name} avatarPath={user.avatar_path} size={72} />
              <div className="absolute inset-0 rounded-full bg-black/0 group-hover:bg-black/40 flex items-center justify-center transition">
                <Camera size={20} className="text-white opacity-0 group-hover:opacity-100" />
              </div>
              <input
                type="file"
                accept="image/*"
                className="hidden"
                disabled={uploadingAvatar}
                onChange={(e) => {
                  const file = e.target.files?.[0]
                  if (file) handleAvatarChange(file)
                  e.target.value = ''
                }}
              />
            </label>
            <div>
              <p className="font-medium text-gray-800">{user.name}</p>
              <p className="text-sm text-gray-500">{user.email}</p>
              {uploadingAvatar && <p className="text-xs text-gray-400 mt-1">Uploading...</p>}
            </div>
          </div>

          <form onSubmit={handleSave} className="space-y-4">
            {error && <p className="text-red-600 text-sm">{error}</p>}
            {success && <p className="text-green-600 text-sm">{success}</p>}

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#D6AE32]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
              <input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="3-30 letters, numbers, or underscores"
                className="w-full border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#D6AE32]"
              />
            </div>

            <button
              type="submit"
              disabled={saving}
              className="bg-[#D6AE32] text-gray-900 rounded px-4 py-2 text-sm font-medium hover:bg-[#B6942B] disabled:opacity-50"
            >
              {saving ? 'Saving...' : 'Save changes'}
            </button>
          </form>
        </div>
      </main>
    </div>
  )
}
