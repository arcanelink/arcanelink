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
  const isInitialized = useAuthStore((state) => state.isInitialized)
  const logout = useAuthStore((state) => state.logout)
  const startSync = useChatStore((state) => state.startSync)
  const stopSync = useChatStore((state) => state.stopSync)
  const loadInitialData = useChatStore((state) => state.loadInitialData)
  const [showCreateRoom, setShowCreateRoom] = useState(false)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Wait for auth initialization
    if (!isInitialized) {
      return
    }

    if (!isAuthenticated) {
      navigate('/login')
      return
    }

    // Load initial data (rooms, etc.) before starting sync
    const initializeChat = async () => {
      try {
        await loadInitialData()
      } catch (error) {
        console.error('Failed to initialize chat:', error)
      } finally {
        setIsLoading(false)
      }
    }

    initializeChat()
    startSync()
    return () => stopSync()
  }, [isAuthenticated, isInitialized, navigate, startSync, stopSync, loadInitialData])

  const handleLogout = () => {
    stopSync()
    logout()
    navigate('/login')
  }

  return (
    <div className="chat-page">
      {!isInitialized || isLoading ? (
        <div className="loading-container">
          <div className="loading-spinner">Loading...</div>
        </div>
      ) : (
        <>
          <Sidebar onCreateRoom={() => setShowCreateRoom(true)} onLogout={handleLogout} />
          <ChatWindow />
          {showCreateRoom && <CreateRoomModal onClose={() => setShowCreateRoom(false)} />}
        </>
      )}
    </div>
  )
}
