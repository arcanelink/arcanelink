// User types
export interface User {
  user_id: string
  username: string
  domain: string
}

// Message types
export interface Message {
  message_id: string
  sender_id: string
  recipient_id?: string
  room_id?: string
  content: MessageContent
  created_at: string
}

export interface MessageContent {
  msgtype: 'm.text' | 'm.image' | 'm.file'
  body: string
  url?: string
}

// Room types
export interface Room {
  room_id: string
  name: string
  creator_id: string
  members: RoomMember[]
  created_at: string
}

export interface RoomMember {
  user_id: string
  role: 'admin' | 'member'
  joined_at: string
}

// Presence types
export interface Presence {
  user_id: string
  status: 'online' | 'offline' | 'away'
  status_msg?: string
  last_active: string
}

// API Request/Response types
export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  domain: string
}

export interface AuthResponse {
  access_token: string
  user_id: string
  expires_in: number
}

export interface SendDirectMessageRequest {
  recipient: string
  content: MessageContent
}

export interface SendRoomMessageRequest {
  room_id: string
  content: MessageContent
}

export interface SyncResponse {
  next_token: string
  messages: Message[]
  presence: Presence[]
  rooms: Room[]
}

export interface CreateRoomRequest {
  name: string
  invite?: string[]
}

export interface CreateRoomResponse {
  room_id: string
}
