// User types
export interface User {
  user_id: string
  username: string
  domain: string
}

// Message types
export interface Message {
  msg_id: string
  sender: string
  recipient?: string
  room_id?: string
  content: MessageContent
  timestamp: number
}

export interface MessageContent {
  msgtype: 'm.text' | 'm.image' | 'm.file' | 'm.audio' | 'm.video'
  body: string
  url?: string
  info?: {
    size?: number
    mimetype?: string
    [key: string]: any
  }
}

export interface RoomEvent {
  event_id: string
  room_id: string
  sender: string
  event_type: string
  content: any
  timestamp: number
}

// Room types
export interface Room {
  room_id: string
  name: string
  creator_id: string
  members: RoomMember[]
  created_at: string
  member_count?: number
  creator?: string
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
  direct_messages?: Message[]
  room_events?: RoomEvent[]
  presence_updates?: Presence[]
}

export interface CreateRoomRequest {
  name: string
  invite?: string[]
}

export interface CreateRoomResponse {
  room_id: string
}
