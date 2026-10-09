'use client'
import { imageSizeError } from "../../utils/images";
import { API, mediaURL } from "../../utils/api";

import { useRealtimeRefresh } from '@/app/webscoket/useRealtimeRefresh'

import { useRouter } from 'next/navigation'
import { useState, useEffect, useRef } from 'react'
import './home.css'
import LikeButton from '../../components/LikeButton'
import ContentActions from '../../components/ContentActions'
import Sidebar from '../../components/Sidebar'
import { redirectBasedOnSession } from '../../utils/session'

interface Author {
  firstName: string
  lastName: string
  nickname: string
  avatar: string
}

interface Comment {
  id: number
  postId: number
  userId: number
  content: string
  image?: string
  createdAt: string
  author: {
    firstName: string
    lastName: string
    nickname?: string
    avatar?: string
  }
}

interface Post {
  id: number
  userId: number
  content: string
  image: string
  privacy: string
  createdAt: string
  author: Author
  likeCount: number
  isLiked: boolean
  comments?: Comment[]
  showComments?: boolean
}

export default function Home() {
  const router = useRouter()
  const [username, setUsername] = useState('')
  const [userId, setUserId] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [posts, setPosts] = useState<Post[]>([])
  const [postsLoading, setPostsLoading] = useState(false)
  const [submittingComment, setSubmittingComment] = useState<{[key: number]: boolean}>({})
  const [newComments, setNewComments] = useState<{[postId: number]: string}>({})
  const [commentImageFiles, setCommentImageFiles] = useState<{[key: number]: File | null}>({});
  const [expandedPosts, setExpandedPosts] = useState<{[postId: number]: boolean}>({});

  const loadedPages = useRef(1)
  const [totalPosts,setTotalPosts] = useState(0)
  const [loadError,setLoadError] = useState('')
  const postsRef = useRef<Post[]>([])
  useEffect(() => { postsRef.current = posts }, [posts])
  useRealtimeRefresh(['posts', 'comments', 'profiles', 'social'], async () => {
    await fetchUserData()
    await fetchPosts(false)
    await Promise.all(postsRef.current.filter(post => post.showComments).map(post => fetchComments(post.id)))
  }, !!userId)


  useEffect(() => {
    redirectBasedOnSession(router, true)
      .then(() => fetchUserData())
      .catch(() => {
        // If there's an error, the redirect will handle it
        setLoading(false)
      })
  }, [router])

  const fetchUserData = async () => {
    try {
      // Fetch user data from the backend
      const response = await fetch(`${API}/profile`, {
        method: 'GET',
        credentials: 'include'
      })

      if (!response.ok) {
        if (response.status === 401) {
          // Unauthorized, redirect to login
          router.push('/')
          return
        }
        throw new Error('Failed to fetch user data')
      }

      const data = await response.json()
      
      // Set the username based on the user's first name or nickname
      if (data.user) {
        setUsername(data.user.nickname || data.user.firstName)
        setUserId(data.user.id)
      }
    } catch (error: any) {
      console.error('Error fetching user data:', error)
      setError(error.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    // Only fetch posts if we have a user ID
    if (userId) {
      fetchPosts()
    }
  }, [userId])

  const fetchPosts = async (showLoading = true) => {
    if (showLoading) setPostsLoading(true)
    try {
      const pages = await Promise.all(Array.from({length:loadedPages.current}, async (_,i) => {
        const response = await fetch(`${API}/posts?page=${i+1}`, {credentials:'include'})
        if (!response.ok) throw new Error('Failed to fetch posts')
        return response.json()
      }))
      const data = {...pages[0], posts:Array.from(new Map(pages.flatMap(page=>page.posts||[]).map((post:Post)=>[post.id,post])).values())}
      setTotalPosts(pages[0].total || 0)
      setLoadError('')
      if (data.success && data.posts) {
        setPosts(previous => data.posts.map((post: Post) => {
          const existing = previous.find(item => item.id === post.id)
          return { ...post, showComments: existing?.showComments ?? false, comments: existing?.comments ?? [] }
        }))
      }
    } catch (error: any) {
      setLoadError(error.message || 'Unable to load posts')
    } finally {
      setPostsLoading(false)
    }
  }

  const fetchComments = async (postId: number) => {
    try {
      console.log(`Fetching comments for post ${postId}`);
      const response = await fetch(`${API}/comments?postId=${postId}`, {
        method: 'GET',
        credentials: 'include'
      });
      
      if (response.status === 404) {
        console.warn(`Post ${postId} not found`);
        return;
      }
      
      if (!response.ok) {
        console.error(`Error response: ${response.status} ${response.statusText}`);
        throw new Error('Failed to fetch comments');
      }
      
      const data = await response.json();
      console.log(`Received comments for post ${postId}:`, data);
      
      if (data.success) {
        setPosts(prevPosts => 
          prevPosts.map(post => 
            post.id === postId ? { ...post, comments: data.comments || [] } : post
          )
        );
      }
    } catch (error) {
      console.error('Error fetching comments:', error);
    }
  };

  const toggleComments = (postId: number) => {
    setPosts(prevPosts => 
      prevPosts.map(post => {
        if (post.id === postId) {
          const newShowComments = !post.showComments;
          
          // If we're showing comments and haven't loaded them yet, fetch them
          if (newShowComments && (!post.comments || post.comments.length === 0)) {
            fetchComments(postId);
          }
          
          return { ...post, showComments: newShowComments };
        }
        return post;
      })
    );
  };

  const handleCommentFileChange = (postId: number, e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const sizeError = imageSizeError(e.target.files[0]);
      if (sizeError) {setError(sizeError);return;}
      setCommentImageFiles(prev => ({
        ...prev,
        [postId]: e.target.files![0]
      }));
    }
  };

  const handleAddComment = async (postId: number) => {
    const commentText = newComments[postId] || ''

    if (commentText.trim() === '' && !commentImageFiles[postId]) {
      return; // Don't submit empty comments without images
    }

    // Validate comment length
    if (commentText.length > 100) {
      console.error('Comment exceeds maximum length of 100 characters')
      return
    }

    setSubmittingComment(prev => ({ ...prev, [postId]: true }));

    try {
      console.log(`Adding comment to post ${postId}: ${newComments[postId]}`);
      console.log('With image file:', commentImageFiles[postId]?.name);
      
      // Create FormData for the request
      const formData = new FormData();
      formData.append('content', newComments[postId] || '');
      
      // Add image if available
      if (commentImageFiles[postId]) {
        formData.append('image', commentImageFiles[postId]);
      }
      
      // Log the form data to verify it's correctly formed
      console.log('Form data entries:');
      for (const [key, value] of formData.entries()) {
        console.log(`${key}: ${value instanceof File ? value.name : value}`);
      }
      
      const response = await fetch(`${API}/comments?postId=${postId}`, {
        method: 'POST',
        credentials: 'include',
        // Don't set Content-Type header manually - let the browser set it with the boundary
        body: formData,
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error(`Error response: ${response.status} ${response.statusText}`, errorText);
        throw new Error('Failed to add comment');
      }
      
      const data = await response.json();
      console.log('Comment submission response:', data);
      
      if (data.success && data.comment) {
        // Update the comments for this post
        setPosts(prevPosts => 
          prevPosts.map(post => {
            if (post.id === postId) {
              const currentComments = post.comments || [];
              return { 
                ...post, 
                comments: [...currentComments, data.comment] 
              };
            }
            return post;
          })
        );
        
        // Clear the new comment input and image
        setNewComments(prev => ({
          ...prev,
          [postId]: ''
        }));
        setCommentImageFiles(prev => ({
          ...prev,
          [postId]: null
        }));
      }
    } catch (error) {
      console.error('Error adding comment:', error);
    } finally {
      setSubmittingComment(prev => ({ ...prev, [postId]: false }));
    }
  };

  const navigateToCreatePost = () => {
    router.push('/create-post')
  }

  const handleDeletePost = async (postId: number) => {
    if (confirm('Are you sure you want to delete this post?')) {
      try {
        const response = await fetch(`${API}/posts?id=${postId}`, {
          method: 'DELETE',
          credentials: 'include',
        })
        
        if (!response.ok) {
          throw new Error('Failed to delete post')
        }
        
        // Remove the deleted post from state
        setPosts(posts.filter(post => post.id !== postId))
      } catch (err: any) {
        console.error('Error deleting post:', err)
      }
    }
  }

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

  const togglePostExpansion = (postId: number) => {
    setExpandedPosts(prev => ({
      ...prev,
      [postId]: !prev[postId]
    }))
  }

  const truncateText = (text: string, maxLength: number = 300) => {
    if (text.length <= maxLength) return text
    return text.substring(0, maxLength) + '...'
  }

  if (loading) return <div className="home-page">Loading...</div>
  if (error) return <div className="home-page">Error: {error}</div>

  return (
    <div className="home-page">
      <Sidebar activePage="home" />
      
      <main className="main-content">
        <div className="dashboard">
          <header className="page-intro">
            <div>
              <span className="eyebrow">The daily loop</span>
              <h1>Good to see you, {username || 'friend'}.</h1>
              <p>Catch up with your circle and share what is happening now.</p>
            </div>
          </header>
          <div className="card feed-card">
            <div className="feed-header">
              <div>
                <span className="section-kicker">Latest from your circle</span>
                <h2 className="card-title">Your feed</h2>
              </div>
              <button 
                className="primary-button create-post-btn" 
                onClick={navigateToCreatePost}
              >
                <span className="create-post-icon">✏️</span>
                Share something
              </button>
            </div>
            
            {loadError && <p role="alert">{loadError}</p>}<div className="posts-container">
              {postsLoading ? (
                <div className="loading-posts">
                  <div className="loading-spinner"></div>
                  <p>Loading posts...</p>
                </div>
              ) : posts.length > 0 ? (
                posts.map(post => (
                  <article key={post.id} className="post-card">
                    <header className="post-header">
                      <div 
                        className="post-author" 
                        onClick={() => router.push(`/profile/${post.userId}`)}
                        role="button"
                        tabIndex={0}
                        style={{ cursor: 'pointer' }}
                      >
                        <div className="author-avatar">
                          {post.author.avatar ? (
                            <img
                              src={mediaURL(post.author.avatar)}
                              alt={`${post.author.firstName}'s avatar`}
                              className="avatar-img"
                            />
                          ) : (
                            <div className="avatar-placeholder">
                              {post.author.firstName.charAt(0)}
                            </div>
                          )}
                        </div>
                        <div className="author-info">
                          <h3 className="author-name">
                            {post.author.nickname || `${post.author.firstName} ${post.author.lastName}`}
                          </h3>
                          <time className="post-date" dateTime={post.createdAt}>
                            {formatDate(post.createdAt)}
                          </time>
                        </div>
                      </div>
                      
                      {post.userId === userId && (
                        <div className="post-actions">
                          <button 
                            onClick={() => router.push(`/edit-post?id=${post.id}`)}
                            className="edit-post-btn"
                            title="Edit post"
                          >
                            <span className="action-icon">✏️</span>
                            Edit
                          </button>
                          <button 
                            onClick={() => handleDeletePost(post.id)}
                            className="delete-post-btn"
                            title="Delete post"
                          >
                            <span className="action-icon">🗑️</span>
                            Delete
                          </button>
                        </div>
                      )}
                    </header>
                    
                    <div className="post-content">
                      {post.content && (
                        <div className="post-text-container">
                          <p className="post-text">
                            {expandedPosts[post.id] || post.content.length <= 300
                              ? post.content
                              : truncateText(post.content, 300)
                            }
                          </p>
                          {post.content.length > 300 && (
                            <button
                              className="read-more-btn"
                              onClick={() => togglePostExpansion(post.id)}
                            >
                              {expandedPosts[post.id] ? 'Show less' : 'Read more'}
                            </button>
                          )}
                        </div>
                      )}
                      {post.image && (
                        <div className="post-image-container">
                          <img
                            src={mediaURL(post.image)}
                            alt="Post image"
                            className="post-image"
                            onError={(e) => {
                              console.error('Failed to load image:', post.image);
                              console.error('Constructed URL:', (e.target as HTMLImageElement).src);
                            }}
                            onLoad={() => {
                              console.log('Successfully loaded image:', post.image);
                            }}
                          />
                        </div>
                      )}
                    </div>
                    
                    <footer className="post-footer">
                      <div className="post-interactions"><LikeButton id={post.id} count={post.likeCount} liked={post.isLiked} /><ContentActions kind="post" id={post.id} reportable={post.userId !== userId} />
                        <button 
                          className={`interaction-btn comments-btn ${post.showComments ? 'active' : ''}`}
                          onClick={() => toggleComments(post.id)}
                        >
                          <span className="interaction-icon">💬</span>
                          <span className="interaction-text">
                            {post.showComments ? 'Hide Comments' : `Comments ${post.comments?.length ? `(${post.comments.length})` : ''}`}
                          </span>
                        </button>
                      </div>
                      
                      {post.showComments && (
                        <div className="comments-section">
                          <div className="comments-header">
                            <h4 className="comments-title">
                              Comments {post.comments?.length ? `(${post.comments.length})` : ''}
                            </h4>
                          </div>
                          
                          <div className="comments-list">
                            {post.comments && post.comments.length > 0 ? (
                              post.comments.map((comment, index) => (
                                <div key={comment.id || index} className="comment-item">
                                  <div className="comment-avatar">
                                    {comment.author?.avatar ? (
                                      <img
                                        src={mediaURL(comment.author.avatar)}
                                        alt={`${comment.author.firstName}'s avatar`}
                                        className="avatar-img"
                                      />
                                    ) : (
                                      <div className="avatar-placeholder">
                                        {comment.author?.firstName?.charAt(0) || 'A'}
                                      </div>
                                    )}
                                  </div>
                                  <div className="comment-content">
                                    <div className="comment-header">
                                      <span className="comment-author">
                                        {comment.author?.nickname || 
                                         `${comment.author?.firstName || 'Anonymous'} ${comment.author?.lastName || ''}`}
                                      </span>
                                      <time className="comment-date" dateTime={comment.createdAt}>
                                        {formatDate(comment.createdAt)}
                                      </time>
                                    </div>
                                    <div className="comment-body">
                                      <p className="comment-text">{comment.content}</p><ContentActions kind="comment" id={comment.id} content={comment.content} canManage={comment.userId === userId} reportable={comment.userId !== userId} />
                                      {comment.image && (
                                        <div className="comment-image-container">
                                          <img
                                            src={mediaURL(comment.image)}
                                            alt="Comment image"
                                            className="comment-image"
                                          />
                                        </div>
                                      )}
                                    </div>
                                  </div>
                                </div>
                              ))
                            ) : (
                              <div className="no-comments">
                                <p>No comments yet. Be the first to comment!</p>
                              </div>
                            )}
                          </div>
                          
                          <div className="comment-form-container">
                            <form onSubmit={(e) => {
                              e.preventDefault();
                              handleAddComment(post.id);
                            }} className="comment-form">
                              <div className="comment-input-container">
                                <div className="current-user-avatar">
                                  <div className="avatar-placeholder">
                                    {username.charAt(0).toUpperCase()}
                                  </div>
                                </div>
                                <div className="comment-input-wrapper">
                                  <textarea
                                    value={newComments[post.id] || ''}
                                    onChange={(e) => setNewComments(prev => ({
                                      ...prev,
                                      [post.id]: e.target.value
                                    }))}
                                    placeholder="Write a comment..."
                                    className="comment-input"
                                    rows={2}
                                    maxLength={100}
                                  />
                                  <div className="comment-input-footer">
                                    <div className="comment-actions">
                                      <input
                                        type="file"
                                        accept="image/*"
                                        style={{ display: 'none' }}
                                        id={`comment-file-${post.id}`}
                                        onChange={(e) => handleCommentFileChange(post.id, e)}
                                      />
                                      <label htmlFor={`comment-file-${post.id}`} className="comment-image-button" title="Add image">
                                        <span className="image-icon">📷</span>
                                      </label>
                                      <div className="comment-char-count">
                                        {(newComments[post.id] || '').length}/100
                                      </div>
                                    </div>
                                    <button 
                                      type="submit" 
                                      className="comment-submit"
                                      disabled={submittingComment[post.id] || (!(newComments[post.id]?.trim()) && !commentImageFiles[post.id])}
                                    >
                                      {submittingComment[post.id] ? (
                                        <span className="submitting">
                                          <span className="spinner"></span>
                                          Posting...
                                        </span>
                                      ) : (
                                        'Post Comment'
                                      )}
                                    </button>
                                  </div>
                                </div>
                              </div>
                              
                              {/* Image preview */}
                              {commentImageFiles[post.id] && (
                                <div className="comment-image-preview">
                                  <div className="preview-container">
                                    <div className="preview-image-wrapper">
                                      <img 
                                        src={URL.createObjectURL(commentImageFiles[post.id]!)} 
                                        alt="Preview" 
                                        className="preview-image" 
                                      />
                                    </div>
                                    <div className="preview-info">
                                      <span className="preview-filename">{commentImageFiles[post.id]?.name}</span>
                                      <button 
                                        type="button" 
                                        className="remove-image-button"
                                        onClick={() => setCommentImageFiles(prev => ({...prev, [post.id]: null}))}
                                        title="Remove image"
                                      >
                                        ✕
                                      </button>
                                    </div>
                                  </div>
                                </div>
                              )}
                            </form>
                          </div>
                        </div>
                      )}
                    </footer>
                  </article>
                ))
              ) : (
                <div className="empty-feed">
                  <div className="empty-feed-icon">📝</div>
                  <h3>Your feed is ready for a first moment</h3>
                  <p>Share an update or find people to begin building your circle.</p>
                  <button 
                    className="primary-button"
                    onClick={navigateToCreatePost}
                  >
                    Create Your First Post
                  </button>
                </div>
              )}
            </div>
            {posts.length < totalPosts && <button className="primary-button" disabled={postsLoading} onClick={()=>{loadedPages.current+=1;fetchPosts()}}>Load older posts</button>}
          </div>
          
        </div>
      </main>
    </div>
  )
}
