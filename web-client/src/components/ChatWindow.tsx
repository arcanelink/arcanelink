import { useState, useEffect, useRef } from 'react'
import { useChatStore } from '../store/chatStore'
import { useAuthStore } from '../store/authStore'
import { apiClient } from '../api/client'
import { MessageItem } from './MessageItem'
import './ChatWindow.css'

export function ChatWindow() {
  const [messageText, setMessageText] = useState('')
  const [sending, setSending] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const currentChat = useChatStore((state) => state.currentChat)
  const messages = useChatStore((state) => state.messages)
  const rooms = useChatStore((state) => state.rooms)
  const removeRoom = useChatStore((state) => state.removeRoom)
  const setCurrentChat = useChatStore((state) => state.setCurrentChat)
  const user = useAuthStore((state) => state.user)

  const filteredMessages = messages.filter((msg) => {
    if (!currentChat) return false

    if (currentChat.type === 'direct') {
      return (
        (msg.sender === user?.user_id && msg.recipient === currentChat.id) ||
        (msg.sender === currentChat.id && msg.recipient === user?.user_id)
      )
    } else {
      return msg.room_id === currentChat.id
    }
  })

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [filteredMessages])

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!messageText.trim() || !currentChat || sending) return

    setSending(true)
    try {
      if (currentChat.type === 'direct') {
        const response = await apiClient.sendDirectMessage({
          recipient: currentChat.id,
          content: { msgtype: 'm.text', body: messageText },
        })

        // Optimistically add the sent message to local state
        const sentMessage = {
          msg_id: response.msg_id,
          sender: user?.user_id || '',
          recipient: currentChat.id,
          content: { msgtype: 'm.text' as const, body: messageText },
          timestamp: response.timestamp,
        }
        useChatStore.getState().addMessage(sentMessage)
      } else {
        await apiClient.sendRoomMessage({
          room_id: currentChat.id,
          content: { msgtype: 'm.text', body: messageText },
        })
      }
      setMessageText('')
    } catch (error) {
      console.error('Failed to send message:', error)
      alert('Failed to send message')
    } finally {
      setSending(false)
    }
  }

  const handleDeleteRoom = async () => {
    if (!currentChat || currentChat.type !== 'room') return

    try {
      await apiClient.deleteRoom(currentChat.id)
      removeRoom(currentChat.id)
      setCurrentChat(null, null)
      setShowDeleteConfirm(false)
      alert('Room deleted successfully')
    } catch (error) {
      console.error('Failed to delete room:', error)
      alert('Failed to delete room. You may not have permission.')
    }
  }

  const currentRoom = currentChat?.type === 'room'
    ? rooms.find(r => r.room_id === currentChat.id)
    : null

  if (!currentChat) {
    return (
      <div className="chat-window">
        <div className="empty-state">
          <h2>Welcome to ArcaneLink</h2>
          <p>Select a conversation or create a new one to start chatting</p>
        </div>
      </div>
    )
  }

  return (
    <div className="chat-window">
      <div className="chat-header">
        <h2>{currentChat.type === 'direct' ? currentChat.id : `# ${currentRoom?.name || currentChat.id}`}</h2>
        {currentChat.type === 'room' && (
          <button
            onClick={() => setShowDeleteConfirm(true)}
            className="btn-delete"
            title="Delete room"
          >
            🗑️
          </button>
        )}
      </div>

      {showDeleteConfirm && (
        <div className="delete-confirm-overlay">
          <div className="delete-confirm-dialog">
            <h3>Delete Room?</h3>
            <p>Are you sure you want to delete this room? This action cannot be undone.</p>
            <div className="delete-confirm-buttons">
              <button onClick={() => setShowDeleteConfirm(false)} className="btn-cancel">
                Cancel
              </button>
              <button onClick={handleDeleteRoom} className="btn-confirm-delete">
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="messages-container">
        {filteredMessages.map((msg) => (
          <MessageItem key={msg.msg_id} message={msg} isOwn={msg.sender === user?.user_id} />
        ))}
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSend} className="message-input-form">
        <input
          type="text"
          value={messageText}
          onChange={(e) => setMessageText(e.target.value)}
          placeholder="Type a message..."
          disabled={sending}
        />
        <button type="submit" disabled={sending || !messageText.trim()}>
          Send
        </button>
      </form>
    </div>
  )
}
