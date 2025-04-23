'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import './home.css'
import Sidebar from '../../components/Sidebar'
import Post from '../../components/Post'
import { WebSocketClient } from '../webscoket/websocket'
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

interface User {
  id: number
  firstName: string
  lastName: string
  email: string
  nickname?: string
}

export default function Home() {
  const router = useRouter()
  const [username, setUsername] = useState('')
  const [userId, setUserId] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [posts, setPosts] = useState<Post[]>([])
  const [postsLoading, setPostsLoading] = useState(false)
  const [wsClient, setWsClient] = useState<WebSocketClient | null>(null)
  const [commentInputs, setCommentInputs] = useState<{[key: number]: string}>({})
  const [submittingComment, setSubmittingComment] = useState<{[key: number]: boolean}>({})
  const [newComments, setNewComments] = useState<{[postId: number]: string}>({})
  const [commentFileInputs, setCommentFileInputs] = useState<{[postId: number]: HTMLInputElement | null}>({});
  const [commentImageFiles, setCommentImageFiles] = useState<{[key: number]: File | null}>({});

  // Initialize WebSocket connection
  useEffect(() => {
    // Only create a WebSocket connection if we have user data
    if (userId) {
        console.log('Getting WebSocket client for user', userId);
        
        const client = WebSocketClient.getInstance();
        
        // Add message handlers for notifications, new posts, etc.
        client.addMessageHandler('private_message', (content) => {
            // Handle incoming private messages
            console.log('Received private message:', content);
            // You could show a notification or update a message counter
        });
        
        client.addMessageHandler('new_post', (content) => {
            // Handle new posts from followed users
            console.log('New post notification:', content);
            // You could refresh the posts or add the new post to the list
            fetchPosts();
        });
        
        // Don't call connect() here - it's handled by getInstance
        
        setWsClient(client);
    }
    
    // No cleanup needed - we want to keep the connection alive
  }, [userId]); // Only depend on userId

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
      const response = await fetch('http://localhost:8080/profile', {
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

  const fetchPosts = async () => {
    setPostsLoading(true)
    try {
      const response = await fetch('http://localhost:8080/posts?page=1', {
        method: 'GET',
        credentials: 'include'
      })

      if (!response.ok) {
        throw new Error('Failed to fetch posts')
      }

      const data = await response.json()
      if (data.success && data.posts) {
        // Initialize posts with showComments property set to false
        const postsWithCommentState = data.posts.map((post: Post) => ({
          ...post,
          showComments: false,
          comments: []
        }))
        setPosts(postsWithCommentState)
      }
    } catch (error: any) {
      console.error('Error fetching posts:', error)
    } finally {
      setPostsLoading(false)
    }
  }

  const fetchComments = async (postId: number) => {
    try {
      console.log(`Fetching comments for post ${postId}`);
      const response = await fetch(`http://localhost:8080/comments?postId=${postId}`, {
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

  const handleCommentChange = (postId: number, value: string) => {
    setCommentInputs(prev => ({ ...prev, [postId]: value }))
  }

  const handleCommentFileChange = (postId: number, e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      console.log('File selected:', e.target.files[0].name);
      setCommentImageFiles(prev => ({
        ...prev,
        [postId]: e.target.files![0]
      }));
    }
  };

  const handleAddComment = async (postId: number) => {
    if ((!newComments[postId] || newComments[postId].trim() === '') && !commentImageFiles[postId]) {
      return; // Don't submit empty comments without images
    }

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
      for (let [key, value] of formData.entries()) {
        console.log(`${key}: ${value instanceof File ? value.name : value}`);
      }
      
      const response = await fetch(`http://localhost:8080/comments?postId=${postId}`, {
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
    }
  };

  const handleLogout = async () => {
    try {
      const response = await fetch('http://localhost:8080/logout', {
        method: 'POST',
        credentials: 'include'
      })
  
      if (response.ok) {
        // Close WebSocket connection before logout
        const client = WebSocketClient.getInstance();
        if (client.socket) {
          client.socket.close(1000, "User logged out");
        }
        
        // Reset the singleton instance using the proper method
        WebSocketClient.resetInstance();
        
        // Clear any stored tokens
        localStorage.removeItem('sessionToken');
        
        router.push('/')
      } else {
        console.error('Error logging out')
      }
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }  

  const navigateToCreatePost = () => {
    router.push('/create-post')
  }

  const handleDeletePost = async (postId: number) => {
    if (confirm('Are you sure you want to delete this post?')) {
      try {
        const response = await fetch(`http://localhost:8080/posts?id=${postId}`, {
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
      day: 'numeric'
    })
  }

  if (loading) return <div className="home-page">Loading...</div>
  if (error) return <div className="home-page">Error: {error}</div>

  return (
    <div className="home-page">
      <Sidebar activePage="home" />
      
      <main className="main-content">
        <div className="dashboard">
          <div className="card feed-card">
            <h2 className="card-title">Your Feed</h2>
            
            <button 
              className="primary-button create-post-btn" 
              onClick={navigateToCreatePost}
            >
              Create New Post
            </button>
            
            <div className="posts-container">
              {postsLoading ? (
                <p className="loading-posts">Loading posts...</p>
              ) : posts.length > 0 ? (
                posts.map(post => (
                  <div key={post.id} className="post-card">
                    <div className="post-header">
                      <div className="post-author">
                        <div className="author-avatar">
                          {post.author.avatar ? (
                            <img src={post.author.avatar} alt={`${post.author.firstName}'s avatar`} />
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
                          <span className="post-date">
                            {formatDate(post.createdAt)}
                          </span>
                        </div>
                      </div>
                      
                      {post.userId === userId && (
                        <div className="post-actions">
                          <button 
                            onClick={() => router.push(`/edit-post?id=${post.id}`)}
                            className="edit-post-btn"
                          >
                            Edit
                          </button>
                          <button 
                            onClick={() => handleDeletePost(post.id)}
                            className="delete-post-btn"
                          >
                            Delete
                          </button>
                        </div>
                      )}
                    </div>
                    
                    <div className="post-content">
                      {post.content && <p className="post-text">{post.content}</p>}
                      {post.image && (
                        <div className="post-image-container">
                          <img 
                            src={post.image.startsWith('http') ? post.image : `http://localhost:8080/${post.image}`} 
                            alt="Post image" 
                            className="post-image" 
                          />
                        </div>
                      )}
                    </div>
                    
                    <div className="post-footer">
                      <div className="post-stats">
                        <span className="like-count">{post.likeCount} likes              </span>
                        <button 
                          className="comments-toggle-btn"
                          onClick={() => toggleComments(post.id)}
                        >
                          {post.showComments ? 'Hide Comments' : 'Show Comments'}
                        </button>
                      </div>
                      
                      {post.showComments && (
                        <div className="comments-section">
                          <h4>Comments</h4>
                          {post.comments && post.comments.map((comment, index) => (
                            <div key={comment.id || index} className="comment">
                              <div className="comment-author">
                                {comment.author?.firstName || 'Anonymous'} {comment.author?.lastName || ''}
                              </div>
                              <div className="comment-content">
                                {comment.content}
                                {comment.image && (
                                  <div className="comment-image-container">
                                    <img 
                                      src={comment.image.startsWith('http') ? comment.image : `http://localhost:8080/${comment.image}`} 
                                      alt="Comment image" 
                                      className="comment-image" 
                                    />
                                  </div>
                                )}
                              </div>
                            </div>
                          ))}
                          
                          <form onSubmit={(e) => {
                            e.preventDefault();
                            handleAddComment(post.id);
                          }} className="comment-form">
                            <input
                              type="text"
                              value={newComments[post.id] || ''}
                              onChange={(e) => setNewComments(prev => ({
                                ...prev,
                                [post.id]: e.target.value
                              }))}
                              placeholder="Write a comment..."
                              className="comment-input"
                            />
                            <div className="comment-form-actions">
                              <input
                                type="file"
                                accept="image/*"
                                style={{ display: 'none' }}
                                id={`comment-file-${post.id}`}
                                onChange={(e) => handleCommentFileChange(post.id, e)}
                              />
                              <label htmlFor={`comment-file-${post.id}`} className="comment-image-button">
                                📷
                              </label>
                              <button type="submit" className="comment-submit">Post</button>
                            </div>
                            
                            {/* Add a visual preview of the selected image */}
                            {commentImageFiles[post.id] && (
                              <div className="comment-image-preview">
                                <div className="preview-image-container">
                                  <img 
                                    src={URL.createObjectURL(commentImageFiles[post.id]!)} 
                                    alt="Preview" 
                                    className="preview-image" 
                                  />
                                </div>
                                <div className="preview-details">
                                  <span className="preview-filename">{commentImageFiles[post.id]?.name}</span>
                                  <button 
                                    type="button" 
                                    className="remove-image-button"
                                    onClick={() => setCommentImageFiles(prev => ({...prev, [post.id]: null}))}
                                  >
                                    ×
                                  </button>
                                </div>
                              </div>
                            )}
                          </form>
                        </div>
                      )}
                    </div>
                  </div>
                ))
              ) : (
                <p className="empty-feed">No posts yet. Start connecting with friends!</p>
              )}
            </div>
          </div>
          
        </div>
      </main>
    </div>
  )
}

