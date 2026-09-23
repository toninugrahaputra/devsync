import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { ConfirmProvider } from './context/ConfirmContext'
import ProtectedRoute from './components/ProtectedRoute'
import AppLayout from './components/AppLayout'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import WorkspacesPage from './pages/WorkspacesPage'
import WorkspaceBoardsPage from './pages/WorkspaceBoardsPage'
import WorkspaceMembersPage from './pages/WorkspaceMembersPage'
import BoardPage from './pages/BoardPage'
import InvitePage from './pages/InvitePage'
import ProfilePage from './pages/ProfilePage'
import AdminUsersPage from './pages/AdminUsersPage'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ConfirmProvider>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/invite/:token" element={<InvitePage />} />
            <Route element={<ProtectedRoute />}>
              <Route element={<AppLayout />}>
                <Route path="/" element={<WorkspacesPage />} />
                <Route path="/w/:workspaceId" element={<WorkspaceBoardsPage />} />
                <Route path="/w/:workspaceId/members" element={<WorkspaceMembersPage />} />
                <Route path="/b/:boardId" element={<BoardPage />} />
                <Route path="/profile" element={<ProfilePage />} />
                <Route path="/admin/users" element={<AdminUsersPage />} />
              </Route>
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </ConfirmProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}
