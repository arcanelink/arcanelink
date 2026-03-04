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
  updatePresence: (presence: Presence[]) => void
  setCurrentChat: (type: 'direct' | 'room', id: string) => void
  startSync: () => void
  stopSync: () => void
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: [],
  rooms: [],
  presence: {},
  currentChat: null,

  addMessage: (message) => {
    set((state) => ({
      messages: [...state.messages, message],
    }))
  },

  addRoom: (room) => {
    set((state) => ({
      rooms: [...state.rooms, room],
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
    set({ currentChat: { type, id } })
  },

  startSync: () => {
    syncManager.start((data) => {
      const state = get()

      // Add new messages
      data.messages.forEach((msg) => state.addMessage(msg))

      // Update presence
      if (data.presence.length > 0) {
        state.updatePresence(data.presence)
      }

      // Update rooms
      data.rooms.forEach((room) => {
        const exists = state.rooms.find((r) => r.room_id === room.room_id)
        if (!exists) {
          state.addRoom(room)
        }
      })
    })
  },

  stopSync: () => {
    syncManager.stop()
  },
}))
