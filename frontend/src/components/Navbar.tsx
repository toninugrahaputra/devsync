import { Link } from 'react-router-dom'
import { LayoutGrid, LogOut, ShieldCheck } from 'lucide-react'
import Avatar from './Avatar'
import { useAuth } from '../context/AuthContext'
import { useConfirm } from '../context/ConfirmContext'

export default function Navbar() {
  const { user, logout } = useAuth()
  const confirm = useConfirm()

  async function handleLogout() {
    if (await confirm({ title: 'Log out', message: 'Are you sure you want to log out?' })) logout()
  }

  return (
    <header className="flex items-center justify-between bg-white border-b border-gray-200 text-gray-800 px-4 py-2">
      <Link to="/" className="flex items-center gap-2 font-semibold">
        <LayoutGrid size={20} className="text-[#D6AE32]" />
        DevSync
      </Link>
      {user && (
        <div className="flex items-center gap-3 text-sm">
          {user.role === 'admin' && (
            <Link to="/admin/users" className="flex items-center gap-1 text-gray-600 hover:text-gray-900" title="Registered users">
              <ShieldCheck size={16} className="text-[#D6AE32]" />
              <span className="hidden sm:inline">Users</span>
            </Link>
          )}
          <Link to="/profile" className="flex items-center gap-2 hover:opacity-80">
            <Avatar name={user.name} avatarPath={user.avatar_path} size={24} />
            <span>{user.name}</span>
          </Link>
          <button onClick={handleLogout} className="flex items-center gap-1 text-gray-500 hover:text-gray-800" title="Logout">
            <LogOut size={16} />
          </button>
        </div>
      )}
    </header>
  )
}
