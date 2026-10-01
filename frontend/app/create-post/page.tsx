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
  const [showUserSelector, setShowUserSelector] = useState(false)
  const [selectedUsers, setSelectedUsers] = useState<number[]>([])
  const [isLoadingUsers, setIsLoadingUsers] = useState(false)
  const [currentUser, setCurrentUser] = useState<any>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Effect to fetch current user profile
  useEffect(() => {
    const fetchCurrentUser = async () => {
      try {
        const response = await fetch('http://localhost:8080/profile', {
          method: 'GET',
          credentials: 'include',
        })

        if (response.ok) {
          const data = await response.json()
          setCurrentUser(data.user)
        }
      } catch (error) {
        console.error('Error fetching current user:', error)
      }
    }

    fetchCurrentUser()
  }, [])

  // Effect to fetch users when privacy is set to private
  useEffect(() => {
    if (privacy === 'private') {
      setShowUserSelector(true)
      fetchFollowers()
    } else {
      setShowUserSelector(false)
      setSelectedUsers([])
    }
  }, [privacy])

  // Function to fetch followers
  const fetchFollowers = async () => {
    setIsLoadingUsers(true)
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
      
      // Now fetch following users (people the current user follows) for private post sharing
      const response = await fetch(`http://localhost:8080/user/${userId}/connections/following`, {
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
        followerId: follower.follower_id, // Current user (who is following)
        followedId: follower.followed_id, // Person being followed (who we want to share with)
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
      setIsLoadingUsers(false)
    }
  }

  // Toggle selection of a user
  const toggleUserSelection = (userId: number) => {
    if (selectedUsers.includes(userId)) {
      setSelectedUsers(selectedUsers.filter(id => id !== userId))
    } else {
      setSelectedUsers([...selectedUsers, userId])
    }
  }

  const validateImage = (file: File): string | null => {
    // Check file type
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif']
    if (!allowedTypes.includes(file.type)) {
      return 'Invalid image format. Only JPEG, PNG, and GIF files are allowed.'
    }

    // Check file size (5MB limit)
    const maxSize = 5 * 1024 * 1024 // 5MB in bytes
    if (file.size > maxSize) {
      return 'Image file size exceeds 5MB limit.'
    }

    return null
  }

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0]

      // Validate the image
      const validationError = validateImage(file)
      if (validationError) {
        setError(validationError)
        e.target.value = '' // Clear the input
        return
      }

      setImage(file)
      setError(null) // Clear any previous errors

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

    // Validate content length
    if (content.length > 100) {
      setError('Post content exceeds maximum length of 100 characters')
      return
    }

    // Validate viewer selection for private posts
    if (privacy === 'private' && selectedUsers.length === 0) {
      setError('Please select at least one user who can view this post')
      return
    }
    
    setIsSubmitting(true)
    setError(null)
    
    try {
      // For private posts with selected viewers, we need to send JSON
      if (privacy === 'private' && selectedUsers.length > 0) {
        const postData = {
          content: content,
          privacy: privacy,
          viewerIds: selectedUsers
        }
        
        if (image) {
          // We need to use FormData for images
          const formData = new FormData()
          formData.append('content', content)
          formData.append('privacy', privacy)
          formData.append('image', image)
          
          // Add viewer IDs as JSON string
          formData.append('viewerIds', JSON.stringify(selectedUsers))
          
          const response = await fetch('http://localhost:8080/posts', {
            method: 'POST',
            credentials: 'include',
            body: formData,
          })
          
          if (!response.ok) {
            const errorText = await response.text()
            throw new Error(`Failed to create post: ${errorText}`)
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
            const errorText = await response.text()
            throw new Error(`Failed to create post: ${errorText}`)
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
          const errorText = await response.text()
          throw new Error(`Failed to create post: ${errorText}`)
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

          {/* Current user profile section */}
          {currentUser && (
            <div className="current-user-section">
              <div className="current-user-avatar">
                {currentUser.avatar ? (
                  <img
                    src={currentUser.avatar.startsWith('http') ? currentUser.avatar : `http://localhost:8080${currentUser.avatar.replace(/\\/g, '/')}`}
                    alt={`${currentUser.firstName}'s avatar`}
                    className="avatar-img"
                  />
                ) : (
                  <div className="avatar-placeholder">
                    {currentUser.firstName?.charAt(0) || 'U'}
                  </div>
                )}
              </div>
              <div className="current-user-info">
                <span className="current-user-name">
                  {currentUser.nickname || `${currentUser.firstName} ${currentUser.lastName}`}
                </span>
                <span className="posting-as">What's on your mind?</span>
              </div>
            </div>
          )}

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
                maxLength={100}
              />
              <div className={`character-counter ${content.length > 100 ? 'over-limit' : ''}`}>
                {content.length}/100 characters
              </div>
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
                <option value="friends">Friends Only</option>
                <option value="private">Private</option>
              </select>
            </div>
            
            {showUserSelector && (
              <div className="form-group follower-selector">
                <label>Select people who can see this post:</label>
                {selectedUsers.length > 0 && (
                  <div className="selected-count">
                    {selectedUsers.length} user{selectedUsers.length !== 1 ? 's' : ''} selected
                  </div>
                )}
                {isLoadingUsers ? (
                  <div className="loading-followers">Loading followers...</div>
                ) : followers.length > 0 ? (
                  <div className="followers-list">
                    {followers.map(follower => (
                      <div 
                        key={follower.id}
                        className={`follower-item ${selectedUsers.includes(follower.followedId) ? 'selected' : ''}`}
                        onClick={() => toggleUserSelection(follower.followedId)}
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
                            checked={selectedUsers.includes(follower.followedId)}
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
