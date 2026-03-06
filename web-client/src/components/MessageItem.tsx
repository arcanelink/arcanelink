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

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
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

  const renderFileMessage = () => {
    const { msgtype, body, url, info } = message.content
    const fileSize = info?.size ? formatFileSize(info.size) : ''

    if (msgtype === 'm.image') {
      return (
        <div className="file-message image-message">
          <img src={url} alt={body} className="message-image" />
          <div className="file-info">
            <span className="file-name">{body}</span>
            {fileSize && <span className="file-size">{fileSize}</span>}
          </div>
        </div>
      )
    }

    if (msgtype === 'm.audio') {
      return (
        <div className="file-message audio-message">
          <audio controls src={url} className="message-audio" />
          <div className="file-info">
            <span className="file-name">🎵 {body}</span>
            {fileSize && <span className="file-size">{fileSize}</span>}
          </div>
        </div>
      )
    }

    if (msgtype === 'm.video') {
      return (
        <div className="file-message video-message">
          <video controls src={url} className="message-video" />
          <div className="file-info">
            <span className="file-name">🎬 {body}</span>
            {fileSize && <span className="file-size">{fileSize}</span>}
          </div>
        </div>
      )
    }

    if (msgtype === 'm.file') {
      return (
        <a href={url} download className="file-message file-download">
          <div className="file-icon">📎</div>
          <div className="file-info">
            <span className="file-name">{body}</span>
            {fileSize && <span className="file-size">{fileSize}</span>}
          </div>
          <div className="download-icon">⬇️</div>
        </a>
      )
    }

    return null
  }

  const isFileMessage = ['m.file', 'm.image', 'm.audio', 'm.video'].includes(message.content.msgtype)
  const urls = isFileMessage ? [] : extractUrls(message.content.body)

  return (
    <div className={`message-item ${isOwn ? 'own' : 'other'}`}>
      <div className="message-bubble">
        {!isOwn && <div className="message-sender">{message.sender}</div>}
        {isFileMessage ? (
          renderFileMessage()
        ) : (
          <>
            <div className="message-content">{renderMessageContent(message.content.body)}</div>
            {urls.length > 0 && (
              <div className="message-previews">
                {urls.slice(0, 3).map((url, index) => (
                  <LinkPreview key={index} url={url} />
                ))}
              </div>
            )}
          </>
        )}
        <div className="message-time">{formatTime(message.timestamp)}</div>
      </div>
    </div>
  )
}
