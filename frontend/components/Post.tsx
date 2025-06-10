'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import './Post.css'

interface Author {
  firstName: string
  lastName: string
  nickname: string
  avatar: string
}

interface PostProps {
  id: number
  userId: number
  content: string
  image: string
  privacy: string
  createdAt: string
  author: Author
  likeCount: number
  isLiked: boolean
  currentUserId: number
  onDelete: (postId: number) => void
}

export default function Post({ 
  id, 
  userId, 
  content, 
  image, 
  privacy, 
  createdAt, 
  author, 
  likeCount, 
  isLiked,
  currentUserId,
  onDelete
}: PostProps) {
  const router = useRouter()
  const [isImageLoaded, setIsImageLoaded] = useState(false)
  const [isLandscape, setIsLandscape] = useState(false)
  const [showFullImage, setShowFullImage] = useState(false)
  
  const isOwner = userId === currentUserId
  
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }
  
  const handleEdit = () => {
    router.push(`/edit-post/${id}`)
  }
  
  const handleDelete = () => {
    if (confirm('Are you sure you want to delete this post?')) {
      onDelete(id)
    }
  }
  
  const handleImageLoad = (e: React.SyntheticEvent<HTMLImageElement>) => {
    const img = e.target as HTMLImageElement
    setIsLandscape(img.naturalWidth > img.naturalHeight)
    setIsImageLoaded(true)
  }
  
  return (
    <div className="post-card">
      <div className="post-header">
        <div className="post-author">
          <div className="author-avatar">
            {author.avatar ? (
              <img src={author.avatar} alt={`${author.firstName}'s avatar`} />
            ) : (
              <div className="avatar-placeholder">
                {author.firstName.charAt(0)}
              </div>
            )}
          </div>
          <div className="author-info">
            <h3 className="author-name">
              {author.nickname || `${author.firstName} ${author.lastName}`}
            </h3>
            <span className="post-date">{formatDate(createdAt)}</span>
          </div>
        </div>
        
        {isOwner && (
          <div className="post-actions">
            <button onClick={handleEdit} className="edit-post-btn">
              Edit
            </button>
            <button onClick={handleDelete} className="delete-post-btn">
              Delete
            </button>
          </div>
        )}
      </div>
      
      <div className="post-content">
        {content && <p className="post-text">{content}</p>}
                  {image && (
                    <>
                      <div 
                        className={`post-image-container ${isImageLoaded ? 'loaded' : 'loading'}`}
                        onClick={() => setShowFullImage(true)}
                      >
                        {!isImageLoaded && <div className="image-loading-spinner"></div>}
                        <img 
                          src={`http://localhost:8080${image}`} 
                          alt="Post content" 
                          className={`post-image ${isLandscape ? 'landscape' : ''}`}
                          onLoad={handleImageLoad}
                        />
                      </div>
            
                      {showFullImage && (
                        <div className="image-modal" onClick={() => setShowFullImage(false)}>
                          <div className="modal-content">
                            <img 
                              src={`http://localhost:8080${image}`} 
                              alt="Full size post content" 
                              className="full-size-image"
                            />
                            <button className="close-modal-btn" onClick={() => setShowFullImage(false)}>
                              ×
                            </button>
                          </div>
                        </div>
                      )}
                    </>
                  )}
                </div>
            <div className="post-footer">
     
      </div>
    </div>
  )
}
