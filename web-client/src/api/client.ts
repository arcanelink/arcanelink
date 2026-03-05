import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  SendDirectMessageRequest,
  SendRoomMessageRequest,
  SyncResponse,
  CreateRoomRequest,
  CreateRoomResponse,
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
    return this.request('/send_direct', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async sendRoomMessage(data: SendRoomMessageRequest): Promise<{ event_id: string }> {
    return this.request('/send_room', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  // Room APIs
  async createRoom(data: CreateRoomRequest): Promise<CreateRoomResponse> {
    return this.request('/rooms/create', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async joinRoom(roomId: string): Promise<void> {
    await this.request(`/rooms/${roomId}/join`, {
      method: 'POST',
    })
  }

  async leaveRoom(roomId: string): Promise<void> {
    await this.request(`/rooms/${roomId}/leave`, {
      method: 'POST',
    })
  }

  async deleteRoom(roomId: string): Promise<void> {
    await this.request('/rooms/delete', {
      method: 'POST',
      body: JSON.stringify({ room_id: roomId }),
    })
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
