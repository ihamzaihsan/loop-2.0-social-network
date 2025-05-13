'use client'

import { useState, useRef, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import './createpost.css'
import Sidebar from '../../components/Sidebar'

// Define the follower interface
interface Follower {
  id: number
  followerId: number
  followedId: number
  status: string
  username: string
  avatar?: string
  selected?: boolean
}

export default function CreatePost() {
  const router = useRouter()
  const [content, setContent] = useState('')
  const [privacy, setPrivacy] = useState('public')
  const [image, setImage] = useState<File | null>(null)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [followers, setFollowers] = useState<Follower[]>([])
  const [showFollowerSelector, setShowFollowerSelector] = useState(false)
  const [selectedFollowers, setSelectedFollowers] = useState<number[]>([])
  const [isLoadingFollowers, setIsLoadingFollowers] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Effect to fetch followers when privacy is set to private
  useEffect(() => {
    if (privacy === 'private') {
      setShowFollowerSelector(true)
      fetchFollowers()
    } else {
      setShowFollowerSelector(false)
      setSelectedFollowers([])
    }
  }, [privacy])

  // Function to fetch followers
  const fetchFollowers = async () => {
    setIsLoadingFollowers(true)
    try {
      // Get the current user ID first (this should be available from the session)
      const userProfileResponse = await fetch('http://localhost:8080/profile', {
        method: 'GET',
        credentials: 'include',
      });
      
      if (!userProfileResponse.ok) {
        throw new Error('Failed to fetch user profile');
      }
      
      const userProfileData = await userProfileResponse.json();
      const userId = userProfileData.user.id;
      
      // Now fetch followers using the user connections endpoint
      const response = await fetch(`http://localhost:8080/user/${userId}/connections/followers`, {
        method: 'GET',
        credentials: 'include',
      })
      
      if (!response.ok) {
        throw new Error('Failed to fetch followers')
      }
      
      const data = await response.json()
      console.log('Followers data:', data); // Debug log
      
      // Process the response data based on its format
      // The user connections endpoint returns connections array
      const followersData = data.connections || [];
      
      // Map the response to our Follower interface
      const mappedFollowers = followersData.map((follower: any) => ({
        id: follower.id,
        followerId: follower.followerId, // Person who follows the user
        followedId: follower.followedId, // User being followed
        status: follower.status,
        username: follower.username || '',
        avatar: follower.avatar,
        selected: false
      }));
      
      setFollowers(mappedFollowers)
    } catch (err: any) {
      console.error('Error fetching followers:', err)
      setError('Failed to load followers. Please try again.')
    } finally {
      setIsLoadingFollowers(false)
    }
  }
  
  // Toggle selection of a follower
  const toggleFollowerSelection = (followerId: number) => {
    if (selectedFollowers.includes(followerId)) {
      setSelectedFollowers(selectedFollowers.filter(id => id !== followerId))
    } else {
      setSelectedFollowers([...selectedFollowers, followerId])
    }
  }

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
    
    // Validate viewer selection for private posts
    if (privacy === 'private' && selectedFollowers.length === 0) {
      setError('Please select at least one follower who can view this post')
      return
    }
    
    setIsSubmitting(true)
    setError(null)
    
    try {
      // For private posts with selected viewers, we need to send JSON
      if (privacy === 'private' && selectedFollowers.length > 0) {
        const postData = {
          content: content,
          privacy: privacy,
          viewerIds: selectedFollowers
        }
        
        if (image) {
          // We need to use FormData for images
          const formData = new FormData()
          formData.append('content', content)
          formData.append('privacy', privacy)
          formData.append('image', image)
          
          // Add viewer IDs as JSON string
          formData.append('viewerIds', JSON.stringify(selectedFollowers))
          
          const response = await fetch('http://localhost:8080/posts', {
            method: 'POST',
            credentials: 'include',
            body: formData,
          })
          
          if (!response.ok) {
            throw new Error('Failed to create post')
          }
        } else {
          // If no image, we can use JSON directly
          const response = await fetch('http://localhost:8080/posts', {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(postData),
          })
          
          if (!response.ok) {
            throw new Error('Failed to create post')
          }
        }
      } else {
        // For public or followers-only posts, use the original FormData approach
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
      }
      
      router.push('/home');
    } catch (err: any) {
      setError(err.message || 'Something went wrong')
    } finally {
      setIsSubmitting(false)
    }
  }
  
  return (
    <div className="create-post-page">
      <Sidebar activePage="" />
      
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
            
            {showFollowerSelector && (
              <div className="form-group follower-selector">
                <label>Select followers who can see this post:</label>
                {isLoadingFollowers ? (
                  <div className="loading-followers">Loading followers...</div>
                ) : followers.length > 0 ? (
                  <div className="followers-list">
                    {followers.map(follower => (
                      <div 
                        key={follower.id} 
                        className={`follower-item ${selectedFollowers.includes(follower.followerId) ? 'selected' : ''}`}
                        onClick={() => toggleFollowerSelection(follower.followerId)}
                      >
                        <div className="follower-avatar">
                          {follower.avatar ? (
                            <img src={follower.avatar} alt={follower.username} />
                          ) : (
                            <div className="avatar-placeholder">{follower.username.charAt(0)}</div>
                          )}
                        </div>
                        <div className="follower-name">{follower.username}</div>
                        <div className="follower-checkbox">
                          <input 
                            type="checkbox" 
                            checked={selectedFollowers.includes(follower.followerId)}
                            onChange={() => {}} // Handled by the div click
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="no-followers">
                    You don't have any followers yet. 
                    {privacy === 'private' && ' Your post will only be visible to you.'}
                  </div>
                )}
              </div>
            )}
            
            <div className="form-actions">
              <button 
                type="button" 
                onClick={() => router.push('/home')}
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
    </div>
  )
}
