import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { useChatStore } from '../store/chatStore'
import { Sidebar } from '../components/Sidebar'
import { ChatWindow } from '../components/ChatWindow'
import { CreateRoomModal } from '../components/CreateRoomModal'
import './ChatPage.css'

export function ChatPage() {
  const navigate = useNavigate()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const logout = useAuthStore((state) => state.logout)
  const startSync = useChatStore((state) => state.startSync)
  const stopSync = useChatStore((state) => state.stopSync)
  const [showCreateRoom, setShowCreateRoom] = useState(false)

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login')
      return
    }

    startSync()
    return () => stopSync()
  }, [isAuthenticated, navigate, startSync, stopSync])

  const handleLogout = () => {
    stopSync()
    logout()
    navigate('/login')
  }

  return (
    <div className="chat-page">
      <Sidebar onCreateRoom={() => setShowCreateRoom(true)} onLogout={handleLogout} />
      <ChatWindow />
      {showCreateRoom && <CreateRoomModal onClose={() => setShowCreateRoom(false)} />}
    </div>
  )
}
