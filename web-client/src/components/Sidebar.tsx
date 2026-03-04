import { useState } from 'react'
import { useChatStore } from '../store/chatStore'
import { useAuthStore } from '../store/authStore'
import './Sidebar.css'

interface SidebarProps {
  onCreateRoom: () => void
  onLogout: () => void
}

export function Sidebar({ onCreateRoom, onLogout }: SidebarProps) {
  const [newChatUserId, setNewChatUserId] = useState('')
  const rooms = useChatStore((state) => state.rooms)
  const messages = useChatStore((state) => state.messages)
  const setCurrentChat = useChatStore((state) => state.setCurrentChat)
  const currentChat = useChatStore((state) => state.currentChat)
  const user = useAuthStore((state) => state.user)

  // Get unique direct message conversations
  const directChats = Array.from(
    new Set(
      messages
        .filter((m) => !m.room_id)
        .map((m) => (m.sender_id === user?.user_id ? m.recipient_id : m.sender_id))
        .filter(Boolean)
    )
  )

  const handleStartDirectChat = (e: React.FormEvent) => {
    e.preventDefault()
    if (newChatUserId.trim()) {
      setCurrentChat('direct', newChatUserId)
      setNewChatUserId('')
    }
  }

  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <h2>ArcaneLink</h2>
        <button onClick={onLogout} className="btn-logout">
          Logout
        </button>
      </div>

      <div className="sidebar-section">
        <h3>Direct Messages</h3>
        <form onSubmit={handleStartDirectChat} className="new-chat-form">
          <input
            type="text"
            value={newChatUserId}
            onChange={(e) => setNewChatUserId(e.target.value)}
            placeholder="@user:domain.com"
          />
          <button type="submit">+</button>
        </form>
        <div className="chat-list">
          {directChats.map((userId) => (
            <div
              key={userId}
              className={`chat-item ${
                currentChat?.type === 'direct' && currentChat.id === userId ? 'active' : ''
              }`}
              onClick={() => setCurrentChat('direct', userId!)}
            >
              <div className="chat-avatar">{userId?.charAt(1).toUpperCase()}</div>
              <div className="chat-info">
                <div className="chat-name">{userId}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="sidebar-section">
        <div className="section-header">
          <h3>Rooms</h3>
          <button onClick={onCreateRoom} className="btn-create">
            +
          </button>
        </div>
        <div className="chat-list">
          {rooms.map((room) => (
            <div
              key={room.room_id}
              className={`chat-item ${
                currentChat?.type === 'room' && currentChat.id === room.room_id ? 'active' : ''
              }`}
              onClick={() => setCurrentChat('room', room.room_id)}
            >
              <div className="chat-avatar">#</div>
              <div className="chat-info">
                <div className="chat-name">{room.name}</div>
                <div className="chat-meta">{room.members.length} members</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
