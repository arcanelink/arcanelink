import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import './LinkPreview.css'

interface LinkPreviewProps {
  url: string
}

interface PreviewData {
  url: string
  title: string
  description: string
  image: string
  site_name: string
}

export function LinkPreview({ url }: LinkPreviewProps) {
  const [preview, setPreview] = useState<PreviewData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    const fetchPreview = async () => {
      try {
        const data = await apiClient.getLinkPreview(url)
        console.log('Link preview data:', data) // Debug log
        setPreview(data)
        setError(false)
      } catch (err) {
        console.error('Failed to fetch link preview for', url, ':', err)
        setError(true)
      } finally {
        setLoading(false)
      }
    }

    fetchPreview()
  }, [url])

  if (loading) {
    return (
      <div className="link-preview loading">
        <div className="link-preview-skeleton">
          <div className="skeleton-image"></div>
          <div className="skeleton-content">
            <div className="skeleton-title"></div>
            <div className="skeleton-description"></div>
          </div>
        </div>
      </div>
    )
  }

  if (error || !preview) {
    return null
  }

  const hasContent = preview.title || preview.description || preview.image

  if (!hasContent) {
    return null
  }

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="link-preview"
    >
      {preview.image && (
        <div className="link-preview-image">
          <img src={preview.image} alt={preview.title || 'Preview'} />
        </div>
      )}
      <div className="link-preview-content">
        {preview.title && (
          <div className="link-preview-title">{preview.title}</div>
        )}
        {preview.description && (
          <div className="link-preview-description">{preview.description}</div>
        )}
        {preview.site_name && (
          <div className="link-preview-site">{preview.site_name}</div>
        )}
        <div className="link-preview-url">{new URL(url).hostname}</div>
      </div>
    </a>
  )
}
