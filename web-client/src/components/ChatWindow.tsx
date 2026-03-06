import { useState, useEffect, useRef } from 'react'
import { useChatStore } from '../store/chatStore'
import { useAuthStore } from '../store/authStore'
import { apiClient } from '../api/client'
import { MessageItem } from './MessageItem'
import { EmojiPicker } from './EmojiPicker'
import './ChatWindow.css'

export function ChatWindow() {
  const [messageText, setMessageText] = useState('')
  const [sending, setSending] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [showMembers, setShowMembers] = useState(false)
  const [showInviteModal, setShowInviteModal] = useState(false)
  const [showEmojiPicker, setShowEmojiPicker] = useState(false)
  const [inviteUserId, setInviteUserId] = useState('')
  const [members, setMembers] = useState<Array<{ user_id: string; joined_at: number }>>([])
  const [loadingMembers, setLoadingMembers] = useState(false)
  const [roomCreator, setRoomCreator] = useState<string | null>(null)
  const [uploadingFile, setUploadingFile] = useState(false)
  const [useMarkdown, setUseMarkdown] = useState(false)
  const [attachedFile, setAttachedFile] = useState<{
    file: File
    fileInfo: {
      file_id: string
      filename: string
      content_type: string
      file_size: number
      url: string
    }
  } | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const currentChat = useChatStore((state) => state.currentChat)
  const messages = useChatStore((state) => state.messages)
  const rooms = useChatStore((state) => state.rooms)
  const removeRoom = useChatStore((state) => state.removeRoom)
  const setCurrentChat = useChatStore((state) => state.setCurrentChat)
  const user = useAuthStore((state) => state.user)

  const filteredMessages = messages
    .filter((msg) => {
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
    .sort((a, b) => a.timestamp - b.timestamp) // Sort by timestamp ascending

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [filteredMessages])

  useEffect(() => {
    // Load room creator when room changes
    if (currentChat?.type === 'room') {
      loadRoomCreator()
    } else {
      setRoomCreator(null)
    }
  }, [currentChat])

  const loadRoomCreator = async () => {
    if (!currentChat || currentChat.type !== 'room') return

    try {
      const state = await apiClient.getRoomState(currentChat.id)
      setRoomCreator(state.creator)
    } catch (error) {
      console.error('Failed to load room state:', error)
    }
  }

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    if ((!messageText.trim() && !attachedFile) || !currentChat || sending) return

    setSending(true)
    try {
      // Determine message type and content
      let msgtype = 'm.text'
      let body = messageText
      let url = undefined
      let info = undefined

      if (attachedFile) {
        // Determine message type based on file type
        if (attachedFile.file.type.startsWith('image/')) {
          msgtype = 'm.image'
        } else if (attachedFile.file.type.startsWith('audio/')) {
          msgtype = 'm.audio'
        } else if (attachedFile.file.type.startsWith('video/')) {
          msgtype = 'm.video'
        } else {
          msgtype = 'm.file'
        }

        // Use user input as body if provided, otherwise use filename
        body = messageText.trim() || attachedFile.fileInfo.filename
        url = attachedFile.fileInfo.url
        info = {
          size: attachedFile.fileInfo.file_size,
          mimetype: attachedFile.fileInfo.content_type,
          filename: attachedFile.fileInfo.filename, // Store original filename
        }
      }

      if (currentChat.type === 'direct') {
        const response = await apiClient.sendDirectMessage({
          recipient: currentChat.id,
          content: {
            msgtype: msgtype as any,
            body,
            format: useMarkdown && !attachedFile ? 'markdown' : undefined,
            url,
            info,
          },
        })

        // Optimistically add the sent message to local state
        const sentMessage = {
          msg_id: response.msg_id,
          sender: user?.user_id || '',
          recipient: currentChat.id,
          content: {
            msgtype: msgtype as any,
            body,
            format: useMarkdown && !attachedFile ? 'markdown' : undefined,
            url,
            info,
          },
          timestamp: response.timestamp,
        }
        useChatStore.getState().addMessage(sentMessage)
      } else {
        const response = await apiClient.sendRoomMessage({
          room_id: currentChat.id,
          content: {
            msgtype: msgtype as any,
            body,
            format: useMarkdown && !attachedFile ? 'markdown' : undefined,
            url,
            info,
          },
        })

        // Optimistically add the sent message to local state
        const sentMessage = {
          msg_id: response.event_id,
          sender: user?.user_id || '',
          room_id: currentChat.id,
          content: {
            msgtype: msgtype as any,
            body,
            format: useMarkdown && !attachedFile ? 'markdown' : undefined,
            url,
            info,
          },
          timestamp: response.timestamp,
        }
        useChatStore.getState().addMessage(sentMessage)
      }
      setMessageText('')
      setAttachedFile(null)
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

  const loadRoomMembers = async () => {
    if (!currentChat || currentChat.type !== 'room') return

    setLoadingMembers(true)
    try {
      const response = await apiClient.getRoomMembers(currentChat.id)
      setMembers(response.members || [])
      setShowMembers(true)
    } catch (error) {
      console.error('Failed to load room members:', error)
      alert('Failed to load room members')
    } finally {
      setLoadingMembers(false)
    }
  }

  const handleInviteUser = async () => {
    if (!currentChat || currentChat.type !== 'room' || !inviteUserId.trim()) return

    try {
      await apiClient.inviteUser(currentChat.id, inviteUserId)
      alert('User invited successfully')
      setInviteUserId('')
      setShowInviteModal(false)
      // Reload members list
      await loadRoomMembers()
    } catch (error: any) {
      console.error('Failed to invite user:', error)
      const errorMessage = error?.message || 'Failed to invite user'
      if (errorMessage.includes('does not exist')) {
        alert('User does not exist. Please check the user ID.')
      } else if (errorMessage.includes('already a member')) {
        alert('User is already a member of this room.')
      } else {
        alert('Failed to invite user: ' + errorMessage)
      }
    }
  }

  const handleLeaveRoom = async () => {
    if (!currentChat || currentChat.type !== 'room') return

    if (!confirm('Are you sure you want to leave this room?')) return

    try {
      await apiClient.leaveRoom(currentChat.id)
      removeRoom(currentChat.id)
      setCurrentChat(null, null)
      alert('Left room successfully')
    } catch (error) {
      console.error('Failed to leave room:', error)
      alert('Failed to leave room')
    }
  }

  const handleEmojiSelect = (emoji: string) => {
    setMessageText((prev) => prev + emoji)
  }

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !currentChat) return

    // Check file size (50MB limit)
    const maxSize = 50 * 1024 * 1024
    if (file.size > maxSize) {
      alert('File size exceeds 50MB limit')
      return
    }

    setUploadingFile(true)
    try {
      // Upload file
      const fileInfo = await apiClient.uploadFile(file)

      // Store file info for later sending
      setAttachedFile({
        file,
        fileInfo,
      })

      // Focus on input for user to add caption
      document.querySelector<HTMLInputElement>('.message-input-form input')?.focus()
    } catch (error) {
      console.error('Failed to upload file:', error)
      alert('Failed to upload file: ' + (error as Error).message)
    } finally {
      setUploadingFile(false)
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  const handleFileButtonClick = () => {
    fileInputRef.current?.click()
  }

  const handleRemoveAttachment = () => {
    setAttachedFile(null)
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
          <div className="room-actions">
            <button
              onClick={loadRoomMembers}
              className="btn-members"
              title="View members"
              disabled={loadingMembers}
            >
              👥 {loadingMembers ? 'Loading...' : 'Members'}
            </button>
            <button
              onClick={() => setShowInviteModal(true)}
              className="btn-invite"
              title="Invite user"
            >
              ➕ Invite
            </button>
            {roomCreator !== user?.user_id && (
              <button
                onClick={handleLeaveRoom}
                className="btn-leave"
                title="Leave room"
              >
                🚪 Leave
              </button>
            )}
            {roomCreator === user?.user_id && (
              <button
                onClick={() => setShowDeleteConfirm(true)}
                className="btn-delete"
                title="Delete room (removes all members)"
              >
                🗑️ Delete
              </button>
            )}
          </div>
        )}
      </div>

      {showInviteModal && (
        <div className="modal-overlay" onClick={() => setShowInviteModal(false)}>
          <div className="modal-dialog" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Invite User to Room</h3>
              <button onClick={() => setShowInviteModal(false)} className="btn-close">×</button>
            </div>
            <div className="modal-body">
              <input
                type="text"
                value={inviteUserId}
                onChange={(e) => setInviteUserId(e.target.value)}
                placeholder="Enter user ID (e.g., @user:domain)"
                className="invite-input"
              />
            </div>
            <div className="modal-footer">
              <button onClick={() => setShowInviteModal(false)} className="btn-cancel">
                Cancel
              </button>
              <button onClick={handleInviteUser} className="btn-confirm" disabled={!inviteUserId.trim()}>
                Invite
              </button>
            </div>
          </div>
        </div>
      )}

      {showMembers && (
        <div className="members-overlay" onClick={() => setShowMembers(false)}>
          <div className="members-dialog" onClick={(e) => e.stopPropagation()}>
            <div className="members-header">
              <h3>Room Members ({members.length})</h3>
              <button onClick={() => setShowMembers(false)} className="btn-close">×</button>
            </div>
            <div className="members-list">
              {members.map((member) => (
                <div key={member.user_id} className="member-item">
                  <div className="member-avatar">{member.user_id.charAt(1).toUpperCase()}</div>
                  <div className="member-info">
                    <div className="member-name">{member.user_id}</div>
                    <div className="member-joined">
                      Joined {new Date(member.joined_at * 1000).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {showDeleteConfirm && (
        <div className="delete-confirm-overlay">
          <div className="delete-confirm-dialog">
            <h3>Delete Room?</h3>
            <p>Are you sure you want to delete this room? All members will be removed and this action cannot be undone.</p>
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
          ref={fileInputRef}
          type="file"
          onChange={handleFileSelect}
          style={{ display: 'none' }}
          disabled={uploadingFile}
        />

        {attachedFile && (
          <div className="file-attachment-preview">
            <div className="attachment-info">
              <span className="attachment-icon">
                {attachedFile.file.type.startsWith('image/') ? '🖼️' :
                 attachedFile.file.type.startsWith('audio/') ? '🎵' :
                 attachedFile.file.type.startsWith('video/') ? '🎬' : '📎'}
              </span>
              <span className="attachment-name">{attachedFile.fileInfo.filename}</span>
              <button
                type="button"
                className="remove-attachment-btn"
                onClick={handleRemoveAttachment}
                title="Remove attachment"
              >
                ✕
              </button>
            </div>
          </div>
        )}

        <div className="input-row">
          <button
            type="button"
            className="file-upload-btn"
            onClick={handleFileButtonClick}
            disabled={uploadingFile || !!attachedFile}
            title="Upload file"
          >
            {uploadingFile ? '⏳' : '📎'}
          </button>
          <button
            type="button"
            className={`markdown-toggle-btn ${useMarkdown ? 'active' : ''}`}
            onClick={() => setUseMarkdown(!useMarkdown)}
            title={useMarkdown ? "Markdown enabled" : "Markdown disabled"}
            disabled={!!attachedFile}
          >
            {useMarkdown ? 'M↓' : 'M'}
          </button>
          <button
            type="button"
            className="emoji-trigger-btn"
            onClick={() => setShowEmojiPicker(!showEmojiPicker)}
            title="Add emoji"
          >
            😀
          </button>
          <input
            type="text"
            value={messageText}
            onChange={(e) => setMessageText(e.target.value)}
            placeholder={attachedFile ? "Add a caption (optional)..." : "Type a message..."}
            disabled={sending || uploadingFile}
          />
          <button type="submit" disabled={sending || (!messageText.trim() && !attachedFile) || uploadingFile}>
            Send
          </button>
        </div>
      </form>

      {showEmojiPicker && (
        <EmojiPicker
          onSelect={handleEmojiSelect}
          onClose={() => setShowEmojiPicker(false)}
        />
      )}
    </div>
  )
}
