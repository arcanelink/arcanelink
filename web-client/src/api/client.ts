import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  SendDirectMessageRequest,
  SendRoomMessageRequest,
  SyncResponse,
  CreateRoomRequest,
  CreateRoomResponse,
  Message,
} from '../types'

const API_BASE = '/_api/v1'

class ApiClient {
  private token: string | null = null

  setToken(token: string) {
    this.token = token
    localStorage.setItem('auth_token', token)
  }

  getToken(): string | null {
    if (!this.token) {
      this.token = localStorage.getItem('auth_token')
    }
    return this.token
  }

  clearToken() {
    this.token = null
    localStorage.removeItem('auth_token')
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.error || `HTTP ${response.status}`)
    }

    return response.json()
  }

  // Auth APIs
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    })
    this.setToken(response.access_token)
    return response
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    })
    this.setToken(response.access_token)
    return response
  }

  // Sync API (Long Polling)
  async sync(since?: string, timeout: number = 30000): Promise<SyncResponse> {
    const params = new URLSearchParams()
    if (since) params.append('since', since)
    params.append('timeout', timeout.toString())

    return this.request<SyncResponse>(`/sync?${params}`)
  }

  // Message APIs
  async sendDirectMessage(data: SendDirectMessageRequest): Promise<{ msg_id: string; timestamp: number }> {
    return this.request('/messages', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async sendRoomMessage(data: SendRoomMessageRequest): Promise<{ event_id: string; timestamp: number }> {
    return this.request(`/rooms/${data.room_id}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content: data.content }),
    })
  }

  async getDirectHistory(peer: string, limit: number = 50): Promise<{
    messages: Message[]
    prev_token: string
    has_more: boolean
  }> {
    const params = new URLSearchParams()
    params.append('peer', peer)
    params.append('limit', limit.toString())
    return this.request(`/messages?${params}`)
  }

  async getRoomHistory(roomId: string, limit: number = 50): Promise<{
    events: Array<{
      event_id: string
      sender: string
      event_type: string
      content: any
      timestamp: number
    }>
    prev_token: string
    has_more: boolean
  }> {
    const params = new URLSearchParams()
    params.append('limit', limit.toString())
    return this.request(`/rooms/${roomId}/messages?${params}`)
  }

  // Room APIs
  async createRoom(data: CreateRoomRequest): Promise<CreateRoomResponse> {
    return this.request('/rooms', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async joinRoom(roomId: string): Promise<void> {
    await this.request(`/rooms/${roomId}/members`, {
      method: 'POST',
    })
  }

  async leaveRoom(roomId: string): Promise<void> {
    await this.request(`/rooms/${roomId}/members`, {
      method: 'DELETE',
    })
  }

  async inviteUser(roomId: string, userId: string): Promise<void> {
    await this.request(`/rooms/${roomId}/members`, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    })
  }

  async deleteRoom(roomId: string): Promise<void> {
    await this.request(`/rooms/${roomId}`, {
      method: 'DELETE',
    })
  }

  async getRooms(): Promise<{ rooms: Array<{ room_id: string; name: string; topic: string; member_count: number }> }> {
    return this.request('/rooms')
  }

  async getRoomMembers(roomId: string): Promise<{ members: Array<{ user_id: string; joined_at: number }> }> {
    return this.request(`/rooms/${roomId}/members`)
  }

  async getRoomState(roomId: string): Promise<{
    room_id: string
    name: string
    topic: string
    creator: string
    created_at: number
    member_count: number
  }> {
    return this.request(`/rooms/${roomId}`)
  }

  async getLinkPreview(url: string): Promise<{
    url: string
    title: string
    description: string
    image: string
    site_name: string
  }> {
    return this.request(`/link_preview?url=${encodeURIComponent(url)}`)
  }

  // File APIs
  async uploadFile(file: File): Promise<{
    file_id: string
    filename: string
    content_type: string
    file_size: number
    url: string
  }> {
    const formData = new FormData()
    formData.append('file', file)

    const headers: Record<string, string> = {}
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${API_BASE}/files`, {
      method: 'POST',
      headers,
      body: formData,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.error || `HTTP ${response.status}`)
    }

    return response.json()
  }

  async getFileInfo(fileId: string): Promise<{
    file_id: string
    filename: string
    content_type: string
    file_size: number
    url: string
  }> {
    return this.request(`/files/${fileId}/info`)
  }

  getFileDownloadUrl(fileId: string): string {
    return `${API_BASE}/files/${fileId}`
  }

  // Presence APIs
  async setPresence(status: 'online' | 'offline' | 'away', statusMsg?: string): Promise<void> {
    await this.request('/presence', {
      method: 'PUT',
      body: JSON.stringify({ status, status_msg: statusMsg }),
    })
  }
}

export const apiClient = new ApiClient()
