import { useState } from 'react'
import { apiClient } from '../api/client'
import { useChatStore } from '../store/chatStore'
import './CreateRoomModal.css'

interface CreateRoomModalProps {
  onClose: () => void
}

export function CreateRoomModal({ onClose }: CreateRoomModalProps) {
  const [roomName, setRoomName] = useState('')
  const [inviteUsers, setInviteUsers] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const addRoom = useChatStore((state) => state.addRoom)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const invite = inviteUsers
        .split(',')
        .map((u) => u.trim())
        .filter(Boolean)

      const response = await apiClient.createRoom({
        name: roomName,
        invite: invite.length > 0 ? invite : undefined,
      })

      // Add room to store (will be updated via sync)
      addRoom({
        room_id: response.room_id,
        name: roomName,
        creator_id: '',
        members: [],
        created_at: new Date().toISOString(),
      })

      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create room')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <h2>Create Room</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="roomName">Room Name</label>
            <input
              id="roomName"
              type="text"
              value={roomName}
              onChange={(e) => setRoomName(e.target.value)}
              placeholder="Enter room name"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="inviteUsers">Invite Users (optional)</label>
            <input
              id="inviteUsers"
              type="text"
              value={inviteUsers}
              onChange={(e) => setInviteUsers(e.target.value)}
              placeholder="@user1:domain.com, @user2:domain.com"
            />
            <small>Comma-separated user IDs</small>
          </div>

          {error && <div className="error-message">{error}</div>}

          <div className="modal-actions">
            <button type="button" onClick={onClose} className="btn-secondary">
              Cancel
            </button>
            <button type="submit" disabled={loading} className="btn-primary">
              {loading ? 'Creating...' : 'Create Room'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
