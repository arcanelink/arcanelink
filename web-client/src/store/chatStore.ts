import { create } from 'zustand'
import type { Message, Room, Presence } from '../types'
import { syncManager } from '../api/sync'

interface ChatState {
  messages: Message[]
  rooms: Room[]
  presence: Record<string, Presence>
  currentChat: { type: 'direct' | 'room'; id: string } | null

  addMessage: (message: Message) => void
  addRoom: (room: Room) => void
  setRooms: (rooms: Room[]) => void
  removeRoom: (roomId: string) => void
  updatePresence: (presence: Presence[]) => void
  setCurrentChat: (type: 'direct' | 'room' | null, id: string | null) => void
  startSync: () => void
  stopSync: () => void
  loadInitialData: () => Promise<void>
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: [],
  rooms: [],
  presence: {},
  currentChat: null,

  addMessage: (message) => {
    set((state) => {
      // Check if message already exists to avoid duplicates
      const exists = state.messages.some((m) => m.msg_id === message.msg_id)
      if (exists) {
        return state
      }
      return {
        messages: [...state.messages, message],
      }
    })
  },

  addRoom: (room) => {
    set((state) => ({
      rooms: [...state.rooms, room],
    }))
  },

  setRooms: (rooms) => {
    set({ rooms })
  },

  removeRoom: (roomId) => {
    set((state) => ({
      rooms: state.rooms.filter((r) => r.room_id !== roomId),
    }))
  },

  updatePresence: (presenceList) => {
    set((state) => {
      const newPresence = { ...state.presence }
      presenceList.forEach((p) => {
        newPresence[p.user_id] = p
      })
      return { presence: newPresence }
    })
  },

  setCurrentChat: (type, id) => {
    if (type === null || id === null) {
      set({ currentChat: null })
    } else {
      set({ currentChat: { type, id } })
    }
  },

  startSync: () => {
    syncManager.start((data) => {
      const state = get()

      // Add new direct messages
      if (data.direct_messages) {
        data.direct_messages.forEach((msg) => state.addMessage(msg))
      }

      // Process room events
      if (data.room_events && data.room_events.length > 0) {
        data.room_events.forEach((event) => {
          // Only process room message events
          if (event.event_type === 'm.room.message') {
            // Convert room event to message format
            const message: Message = {
              msg_id: event.event_id,
              sender: event.sender,
              room_id: event.room_id,
              content: {
                msgtype: event.content.msgtype || 'm.text',
                body: event.content.body || '',
                url: event.content.url,
              },
              timestamp: event.timestamp,
            }
            state.addMessage(message)
          }
        })
      }

      // Update presence
      if (data.presence_updates && data.presence_updates.length > 0) {
        state.updatePresence(data.presence_updates)
      }
    })
  },

  stopSync: () => {
    syncManager.stop()
  },

  loadInitialData: async () => {
    try {
      const { apiClient } = await import('../api/client')

      // Load rooms
      const roomsData = await apiClient.getRooms()
      if (roomsData.rooms && roomsData.rooms.length > 0) {
        const rooms: Room[] = roomsData.rooms.map(r => ({
          room_id: r.room_id,
          name: r.name,
          creator_id: '',
          members: [],
          created_at: '',
        }))
        set({ rooms })
      }

      // Load recent messages to discover conversations
      const syncData = await apiClient.sync(undefined, 0)

      if (syncData.direct_messages && syncData.direct_messages.length > 0) {
        // Add initial messages
        const state = get()
        syncData.direct_messages.forEach(msg => state.addMessage(msg))

        // Get current user ID from auth store
        const { useAuthStore } = await import('./authStore')
        const currentUserId = useAuthStore.getState().user?.user_id

        if (currentUserId) {
          // Extract unique conversation peers
          const peers = new Set<string>()
          syncData.direct_messages.forEach(msg => {
            const peer = msg.sender === currentUserId ? msg.recipient : msg.sender
            if (peer) peers.add(peer)
          })

          // Load history for each peer
          for (const peer of peers) {
            try {
              const history = await apiClient.getDirectHistory(peer, 50)
              if (history.messages) {
                history.messages.forEach(msg => state.addMessage(msg))
              }
            } catch (error) {
              console.error(`Failed to load history for ${peer}:`, error)
            }
          }
        }
      }
    } catch (error) {
      console.error('Failed to load initial data:', error)
    }
  },
}))
