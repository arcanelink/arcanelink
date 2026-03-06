import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface SecureImageProps {
  src: string
  alt: string
  className?: string
}

export function SecureImage({ src, alt, className }: SecureImageProps) {
  const [blobUrl, setBlobUrl] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    let objectUrl: string | null = null

    const loadImage = async () => {
      try {
        setLoading(true)
        setError(false)

        const token = apiClient.getToken()
        if (!token) {
          setError(true)
          return
        }

        // Fetch image with authentication
        const response = await fetch(src, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })

        if (!response.ok) {
          throw new Error('Failed to load image')
        }

        const blob = await response.blob()
        objectUrl = URL.createObjectURL(blob)
        setBlobUrl(objectUrl)
      } catch (err) {
        console.error('Failed to load image:', err)
        setError(true)
      } finally {
        setLoading(false)
      }
    }

    loadImage()

    // Cleanup
    return () => {
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl)
      }
    }
  }, [src])

  if (loading) {
    return (
      <div className={`${className} secure-image-loading`}>
        <div className="loading-spinner">Loading...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={`${className} secure-image-error`}>
        <div className="error-message">Failed to load image</div>
      </div>
    )
  }

  return <img src={blobUrl} alt={alt} className={className} />
}
