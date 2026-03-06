import { useState } from 'react'
import type { Message } from '../types'
import { LinkPreview } from './LinkPreview'
import { SecureImage } from './SecureImage'
import { SecureMedia } from './SecureMedia'
import { MarkdownRenderer } from './MarkdownRenderer'
import './MessageItem.css'
import './SecureImage.css'
import './MarkdownRenderer.css'

interface MessageItemProps {
  message: Message
  isOwn: boolean
}

export function MessageItem({ message, isOwn }: MessageItemProps) {
  const [showRawMessage, setShowRawMessage] = useState(false)

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

  const renderMessageContent = (text: string, format?: string) => {
    // If format is specified as markdown, use MarkdownRenderer
    if (format === 'markdown') {
      return <MarkdownRenderer content={text} />
    }

    // Default: plain text with clickable links
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

    // Get original filename from info or body
    // If body looks like a filename (has extension), use it
    // Otherwise, body is the caption/description
    const hasExtension = body && /\.[a-zA-Z0-9]+$/.test(body)
    const filename = hasExtension ? body : (info?.filename || body)
    const caption = hasExtension ? null : body

    if (msgtype === 'm.image') {
      return (
        <div className="file-message image-message">
          <SecureImage src={url!} alt={filename} className="message-image" />
          {caption && (
            <div className="file-caption">{caption}</div>
          )}
          <div className="file-info">
            <span className="file-name">{filename}</span>
            {fileSize && <span className="file-size">{fileSize}</span>}
          </div>
        </div>
      )
    }

    if (msgtype === 'm.audio') {
      return (
        <div className="file-message audio-message">
          <SecureMedia src={url!} type="audio" className="message-audio" />
          {caption && (
            <div className="file-caption">{caption}</div>
          )}
          <div className="file-info">
            <span className="file-name">🎵 {filename}</span>
            {fileSize && <span className="file-size">{fileSize}</span>}
          </div>
        </div>
      )
    }

    if (msgtype === 'm.video') {
      return (
        <div className="file-message video-message">
          <SecureMedia src={url!} type="video" className="message-video" />
          {caption && (
            <div className="file-caption">{caption}</div>
          )}
          <div className="file-info">
            <span className="file-name">🎬 {filename}</span>
            {fileSize && <span className="file-size">{fileSize}</span>}
          </div>
        </div>
      )
    }

    if (msgtype === 'm.file') {
      return (
        <div className="file-message">
          <a href={url} download className="file-download">
            <div className="file-icon">📎</div>
            <div className="file-info">
              <span className="file-name">{filename}</span>
              {fileSize && <span className="file-size">{fileSize}</span>}
            </div>
            <div className="download-icon">⬇️</div>
          </a>
          {caption && (
            <div className="file-caption">{caption}</div>
          )}
        </div>
      )
    }

    return null
  }

  const isFileMessage = ['m.file', 'm.image', 'm.audio', 'm.video'].includes(message.content.msgtype)
  const urls = isFileMessage ? [] : extractUrls(message.content.body)

  const toggleRawMessage = () => {
    setShowRawMessage(!showRawMessage)
  }

  const copyRawMessage = () => {
    const rawJson = JSON.stringify(message, null, 2)
    navigator.clipboard.writeText(rawJson)
      .then(() => alert('Raw message copied to clipboard!'))
      .catch(() => alert('Failed to copy'))
  }

  return (
    <div className={`message-item ${isOwn ? 'own' : 'other'}`}>
      <div className="message-bubble">
        {!isOwn && <div className="message-sender">{message.sender}</div>}
        {isFileMessage ? (
          renderFileMessage()
        ) : (
          <>
            <div className="message-content">
              {renderMessageContent(message.content.body, message.content.format)}
            </div>
            {urls.length > 0 && (
              <div className="message-previews">
                {urls.slice(0, 3).map((url, index) => (
                  <LinkPreview key={index} url={url} />
                ))}
              </div>
            )}
          </>
        )}

        <div className="message-footer">
          <div className="message-time">{formatTime(message.timestamp)}</div>
          <button
            className="raw-message-toggle"
            onClick={toggleRawMessage}
            title="Show raw message"
          >
            {showRawMessage ? '📋' : '🔍'}
          </button>
        </div>

        {showRawMessage && (
          <div className="raw-message-container">
            <div className="raw-message-header">
              <span>Raw Message</span>
              <button
                className="copy-raw-btn"
                onClick={copyRawMessage}
                title="Copy to clipboard"
              >
                📋 Copy
              </button>
            </div>
            <pre className="raw-message-content">
              {JSON.stringify(message, null, 2)}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
