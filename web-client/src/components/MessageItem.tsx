import type { Message } from '../types'
import './MessageItem.css'

interface MessageItemProps {
  message: Message
  isOwn: boolean
}

export function MessageItem({ message, isOwn }: MessageItemProps) {
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className={`message-item ${isOwn ? 'own' : 'other'}`}>
      <div className="message-bubble">
        {!isOwn && <div className="message-sender">{message.sender_id}</div>}
        <div className="message-content">{message.content.body}</div>
        <div className="message-time">{formatTime(message.created_at)}</div>
      </div>
    </div>
  )
}
