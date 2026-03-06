import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface SecureMediaProps {
  src: string
  type: 'audio' | 'video'
  className?: string
}

export function SecureMedia({ src, type, className }: SecureMediaProps) {
  const [blobUrl, setBlobUrl] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    let objectUrl: string | null = null

    const loadMedia = async () => {
      try {
        setLoading(true)
        setError(false)

        const token = apiClient.getToken()
        if (!token) {
          setError(true)
          return
        }

        // Fetch media with authentication
        const response = await fetch(src, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })

        if (!response.ok) {
          throw new Error('Failed to load media')
        }

        const blob = await response.blob()
        objectUrl = URL.createObjectURL(blob)
        setBlobUrl(objectUrl)
      } catch (err) {
        console.error('Failed to load media:', err)
        setError(true)
      } finally {
        setLoading(false)
      }
    }

    loadMedia()

    // Cleanup
    return () => {
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl)
      }
    }
  }, [src])

  if (loading) {
    return (
      <div className={`${className} secure-media-loading`}>
        <div className="loading-spinner">Loading {type}...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={`${className} secure-media-error`}>
        <div className="error-message">Failed to load {type}</div>
      </div>
    )
  }

  if (type === 'audio') {
    return <audio controls src={blobUrl} className={className} />
  }

  return <video controls src={blobUrl} className={className} />
}
