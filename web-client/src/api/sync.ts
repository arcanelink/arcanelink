import { apiClient } from './client'
import type { SyncResponse } from '../types'

export class SyncManager {
  private isRunning = false
  private nextToken: string | null = null
  private onSyncCallback: ((data: SyncResponse) => void) | null = null

  start(onSync: (data: SyncResponse) => void) {
    this.onSyncCallback = onSync
    this.isRunning = true
    this.poll()
  }

  stop() {
    this.isRunning = false
  }

  private async poll() {
    while (this.isRunning) {
      try {
        const data = await apiClient.sync(this.nextToken || undefined, 30000)
        this.nextToken = data.next_token

        if (this.onSyncCallback) {
          this.onSyncCallback(data)
        }
      } catch (error) {
        console.error('Sync error:', error)
        // Wait before retrying on error
        await new Promise(resolve => setTimeout(resolve, 5000))
      }
    }
  }
}

export const syncManager = new SyncManager()
