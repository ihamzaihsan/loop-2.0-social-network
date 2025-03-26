'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import './home.css'
import Sidebar from '../../components/Sidebar'
import Post from '../../components/Post'
import { WebSocketClient } from '../webscoket/websocket'

interface Author {
  firstName: string
  lastName: string
  nickname: string
  avatar: string
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

    fetchUserData()
  }, [router])

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
        setPosts(data.posts)
      }
    } catch (error: any) {
      console.error('Error fetching posts:', error)
    } finally {
      setPostsLoading(false)
    }
  }

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
                            {new Date(post.createdAt).toLocaleDateString('en-US', {
                              year: 'numeric',
                              month: 'short',
                              day: 'numeric'
                            })}
                          </span>
                        </div>
                      </div>
                      
                      {post.userId === userId && (
                        <div className="post-actions">
                          <button 
                            onClick={() => router.push(`/edit-post/${post.id}`)}
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
                          <img src={post.image} alt="Post image" className="post-image" />
                        </div>
                      )}
                    </div>
                    
                    <div className="post-footer">
                      <div className="post-stats">
                        <span className="like-count">{post.likeCount} likes</span>
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <p className="empty-feed">No posts yet. Start connecting with friends!</p>
              )}
            </div>
          </div>
          
          <div className="card friends-card">
            <h2 className="card-title">Friends</h2>
            <p className="empty-friends">Connect with new friends to see them here.</p>
            <button 
              className="secondary-button find-friends-btn" 
              onClick={() => router.push('/find-friends')}
            >
              Find Friends
            </button>
          </div>
        </div>
      </main>
    </div>
  )
}