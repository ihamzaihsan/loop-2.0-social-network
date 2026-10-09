'use client'
import { API } from "../../utils/api";

import { useRealtimeRefresh } from '@/app/webscoket/useRealtimeRefresh'

import { useRouter, useSearchParams } from 'next/navigation'
import { useState, useEffect, Suspense } from 'react'
import './edit-post.css'
import Sidebar from '../../components/Sidebar'
import { redirectBasedOnSession } from '../../utils/session'

interface Post {
  id: number
  userId: number
  content: string
  image: string
  privacy: string
}

function EditPostContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const postId = searchParams.get('id')
  const [post, setPost] = useState<Post | null>(null)
  const [content, setContent] = useState('')
  const [privacy, setPrivacy] = useState('public')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [remoteChange, setRemoteChange] = useState(false)

  useEffect(() => {
    if (!postId) {
      setError('No post ID provided')
      setLoading(false)
      return
    }

    redirectBasedOnSession(router, true)
      .then(() => fetchPost())
      .catch(() => {
        // If there's an error, the redirect will handle it
        setLoading(false)
      })
  }, [router, postId])

  useRealtimeRefresh(['posts'], () => fetchPost(true), !!postId)

  const fetchPost = async (preserveDraft = false) => {
    if (!postId) return

    try {
      const response = await fetch(`${API}/posts?id=${postId}`, {
        method: 'GET',
        credentials: 'include'
      })

      if (!response.ok) {
        if (response.status === 401) {
          router.push('/')
          return
        }
        if (response.status === 404) {
          setError('This post has been deleted.')
          return
        }
        throw new Error('Failed to fetch post')
      }

      const data = await response.json()
      if (data.success && data.post) {
        if (preserveDraft && post && (content !== post.content || privacy !== post.privacy)) {
          setRemoteChange(data.post.content !== post.content || data.post.privacy !== post.privacy)
          return
        }
        setPost(data.post)
        setContent(data.post.content)
        setPrivacy(data.post.privacy)
        setRemoteChange(false)
      } else {
        throw new Error('Post not found')
      }
    } catch (error: any) {
      console.error('Error fetching post:', error)
      setError(error.message)
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!postId || submitting) return
    setSubmitting(true)
    
    try {
      const response = await fetch(`${API}/posts?id=${postId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({
          content,
          privacy
        })
      })

      if (!response.ok) {
        throw new Error('Failed to update post')
      }

      const data = await response.json()
      if (data.success) {
        router.push('/home')
      } else {
        throw new Error(data.message || 'Failed to update post')
      }
    } catch (error: any) {
      console.error('Error updating post:', error)
      setError(error.message)
    } finally {
      setSubmitting(false)
    }
  }

  const handleCancel = () => {
    router.push('/home')
  }

  if (loading) return <div className="edit-post-page">Loading...</div>
  if (error) return <div className="edit-post-page">Error: {error}</div>
  if (!post) return <div className="edit-post-page">Post not found</div>

  return (
    <div className="edit-post-page">
      <Sidebar activePage="home" />
      
      <main className="main-content">
        <div className="edit-post-container">
          <div className="card edit-post-card">
            {remoteChange && <p role="status">This post changed in another tab. Your draft is preserved. <button type="button" onClick={() => fetchPost()}>Load saved version</button></p>}
            <h2 className="card-title">Edit Post</h2>
            
            <form onSubmit={handleSubmit} className="edit-post-form">
              <div className="form-group">
                <label htmlFor="content">Content</label>
                <textarea
                  id="content"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="What's on your mind?"
                  required
                />
              </div>
              
              <div className="form-group">
                <label htmlFor="privacy">Privacy</label>
                <select
                  id="privacy"
                  value={privacy}
                  onChange={(e) => setPrivacy(e.target.value)}
                >
                  <option value="public">Public</option>
                  <option value="private">Private</option>
                  <option value="friends">Friends Only</option>
                </select>
              </div>
              
              {post.image && (
                <div className="current-image">
                  <p>Current Image:</p>
                  <img src={post.image} alt="Post image" className="post-image-preview" />
                  <p className="image-note">Note: Images cannot be changed after posting.</p>
                </div>
              )}
              
              <div className="form-actions">
                <button 
                  type="button" 
                  onClick={handleCancel}
                  className="cancel-button"
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="save-button"
                  disabled={submitting}
                >
                  {submitting ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        </div>
      </main>
    </div>
  )
}

export default function EditPost() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <EditPostContent />
    </Suspense>
  )
}
