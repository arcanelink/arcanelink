import type { Message } from '../types'
import { LinkPreview } from './LinkPreview'
import './MessageItem.css'

interface MessageItemProps {
  message: Message
  isOwn: boolean
}

export function MessageItem({ message, isOwn }: MessageItemProps) {
  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  const extractUrls = (text: string): string[] => {
    const urlPattern = /(https?:\/\/[^\s]+)/g
    const matches = text.match(urlPattern)
    return matches || []
  }

  const renderMessageContent = (text: string) => {
    const urlPattern = /(https?:\/\/[^\s]+)/g
    const parts = text.split(urlPattern)

    return parts.map((part, index) => {
      if (part.match(urlPattern)) {
        return (
          <a
            key={index}
            href={part}
            target="_blank"
            rel="noopener noreferrer"
            className="message-link"
          >
            {part}
          </a>
        )
      }
      return <span key={index}>{part}</span>
    })
  }

  const urls = extractUrls(message.content.body)

  return (
    <div className={`message-item ${isOwn ? 'own' : 'other'}`}>
      <div className="message-bubble">
        {!isOwn && <div className="message-sender">{message.sender}</div>}
        <div className="message-content">{renderMessageContent(message.content.body)}</div>
        {urls.length > 0 && (
          <div className="message-previews">
            {urls.slice(0, 3).map((url, index) => (
              <LinkPreview key={index} url={url} />
            ))}
          </div>
        )}
        <div className="message-time">{formatTime(message.timestamp)}</div>
      </div>
    </div>
  )
}
