'use client'

import { useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import './createpost.css'

export default function CreatePost() {
  const router = useRouter()
  const [content, setContent] = useState('')
  const [privacy, setPrivacy] = useState('public')
  const [image, setImage] = useState<File | null>(null)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0]
      setImage(file)
      
      // Create preview
      const reader = new FileReader()
      reader.onload = (e) => {
        setImagePreview(e.target?.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

  const handleRemoveImage = () => {
    setImage(null)
    setImagePreview(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!content.trim() && !image) {
      setError('Please add some content or an image to your post')
      return
    }
    
    setIsSubmitting(true)
    setError(null)
    
    try {
      const formData = new FormData()
      formData.append('content', content)
      formData.append('privacy', privacy)
      if (image) {
        formData.append('image', image)
      }
      
      const response = await fetch('http://localhost:8080/posts', {
        method: 'POST',
        credentials: 'include',
        body: formData,
      })
      
      if (!response.ok) {
        throw new Error('Failed to create post')
      }
      
    
      router.push('/home');
    } catch (err: any) {
      setError(err.message || 'Something went wrong')
    } finally {
      setIsSubmitting(false)
    }
  }
  
  return (
    <div className="create-post-container">
      <div className="create-post-card">
        <h1 className="create-post-title">Create a New Post</h1>
        
        {error && <div className="error-message">{error}</div>}
        
        <form onSubmit={handleSubmit} className="create-post-form">
          <div className="form-group">
            <label htmlFor="content">What's on your mind?</label>
            <textarea
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Share your thoughts..."
              rows={5}
              className="content-textarea"
            />
          </div>
          
          {imagePreview && (
            <div className="image-preview-container">
              <img src={imagePreview} alt="Preview" className="image-preview" />
              <button 
                type="button" 
                onClick={handleRemoveImage}
                className="remove-image-btn"
              >
                Remove Image
              </button>
            </div>
          )}
          
          <div className="form-group">
            <label htmlFor="image">Add an Image</label>
            <input
              type="file"
              id="image"
              ref={fileInputRef}
              onChange={handleImageChange}
              accept="image/*"
              className="file-input"
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="privacy">Privacy Setting</label>
            <select
              id="privacy"
              value={privacy}
              onChange={(e) => setPrivacy(e.target.value)}
              className="privacy-select"
            >
              <option value="public">Public</option>
              <option value="almost_private">Followers Only</option>
              <option value="private">Private</option>
            </select>
          </div>
          
          <div className="form-actions">
            <button 
              type="button" 
              onClick={() => router.back()}
              className="cancel-button"
            >
              Cancel
            </button>
            <button 
              type="submit" 
              disabled={isSubmitting}
              className="submit-button"
            >
              {isSubmitting ? 'Posting...' : 'Create Post'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
