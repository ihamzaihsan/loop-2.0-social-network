'use client'

import { useRealtimeRefresh } from '@/app/webscoket/useRealtimeRefresh'

import { useRouter, useParams } from 'next/navigation'
import { useState, useEffect } from 'react'
import '../../home/home.css'
import '../profile.css'
import { mediaURL } from '@/utils/api'
import LikeButton from '@/components/LikeButton'
import ContentActions from '@/components/ContentActions'
import UserSafetyActions from '@/components/UserSafetyActions'
import Sidebar from '../../../components/Sidebar'

interface User {
  id: number
  firstName?: string
  lastName?: string
  nickname?: string
  aboutMe?: string
  email?: string
  avatar?: string
  isPrivate?: boolean 
  createdAt?: string
}

interface Post {
  id: number
  userId: number
  content: string
  image: string
  privacy: string
  createdAt: string
  author: {
    firstName: string
    lastName: string
    nickname?: string
    avatar?: string
  }
  likeCount: number
  isLiked: boolean
}

interface ProfileData {
  user: User
  posts: Post[]
  postsCount: number
  followersCount: number
  followingCount: number
  followers: User[]
  following: User[]
  isCurrentUser: boolean
  isFollowing: boolean
  pendingFollow?: boolean
  isPrivate?: boolean
}

export default function UserProfile() {
  const router = useRouter()
  const params = useParams()
  const userId = params.id as string
  const [profile, setProfile] = useState<ProfileData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isFollowLoading, setIsFollowLoading] = useState(false)

  // Add state for modals
  const [activeModal, setActiveModal] = useState<'followers' | 'following' | null>(null);
  const loadingUserDetails = false;

  // Add modal handlers
  const openFollowersModal = () => setActiveModal('followers');
  const openFollowingModal = () => setActiveModal('following');
  const closeModal = () => setActiveModal(null);

  useEffect(() => {
    if (userId) {
      fetchProfile()
    }
  }, [userId])

  const fetchProfile = async () => {
    try {
      const response = await fetch(`http://localhost:8080/profile/${userId}`, {
        method: 'GET',
        credentials: 'include'
      })

      if (!response.ok) {
        if (response.status === 401) {
          router.push('/')
          return
        }
        if (response.status === 404) {
          setError('User not found')
          return
        }
        throw new Error('Failed to fetch profile data')
      }

      const data = await response.json()
      console.log('Profile data:', data)
      setProfile(data)
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useRealtimeRefresh(['profiles', 'posts', 'social'], fetchProfile, !!userId)

  const handleFollowToggle = async () => {
    if (!profile || profile.isCurrentUser) return
    
    setIsFollowLoading(true)
    
    try {
      const endpoint = profile.isFollowing ? '/unfollow' : '/follow'
      const response = await fetch(`http://localhost:8080${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          followed_id: profile.user.id
        }),
        credentials: 'include'
      })
      
      if (!response.ok) {
        throw new Error('Failed to update follow status')
      }
      
      await fetchProfile()
    } catch (err: any) {
      console.error('Error updating follow status:', err)
      setError(err.message)
    } finally {
      setIsFollowLoading(false)
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

  const navigateToChat = () => {
    if (profile && !profile.isCurrentUser) {
      router.push(`/chat?user=${profile.user.id}`)
    }
  }

  const navigateToProfile = (userId: number) => {
    router.push(`/profile/${userId}`);
    closeModal();
  };

  if (loading) return <div className="profile-page">Loading profile...</div>
  if (error) return <div className="profile-page">Error: {error}</div>

  return (
    <div className="profile-page">
      <Sidebar activePage="profile" />
      
      <main className="main-content">{profile && !profile.isCurrentUser && <UserSafetyActions id={profile.user.id} name={profile.user.firstName || "this user"} />}
        {profile && (
          <div className="profile-container">
            <div className="card profile-card">
              <div className="profile-header">
                <div className="profile-avatar">
                  {profile.user.avatar ? (
                    <img
                      src={`http://localhost:8080${encodeURI(profile.user.avatar)}`}
                      alt={`${profile.user.firstName}'s avatar`}
                      className="avatar-image"
                    />
                  ) : (
                    <div className="avatar-placeholder">
                      {profile.user.firstName?.charAt(0) || '?'}
                    </div>
                  )}
                </div>
                
                <div className="profile-info">
                  <h2 className="profile-name">
                    {profile.user.firstName} {profile.user.lastName}
                  </h2>
                  <p className="profile-nickname">@{profile.user.nickname || profile.user.firstName?.toLowerCase()}</p>
                  
                  {!profile.isCurrentUser && (
                    <div className="profile-actions">
                      <button 
                        className={`follow-button ${profile.isFollowing ? 'following' : 'not-following'}`}
                        onClick={handleFollowToggle}
                        disabled={isFollowLoading || profile.pendingFollow}
                      >
                        {isFollowLoading ? 'Loading...' : profile.isFollowing ? 'Unfollow' : profile.pendingFollow ? 'Pending' : 'Follow'}
                      </button>
                      
                      {profile.isFollowing && (
                        <button 
                          className="message-button"
                          onClick={navigateToChat}
                        >
                          Message
                        </button>
                      )}
                    </div>
                  )}
                  
                  {profile.user.isPrivate && (
                    <div className="privacy-controls">
                      <span className="privacy-badge private">
                        🔒 Private Account
                      </span>
                    </div>
                  )}
                </div>
              </div>
              
              {profile.user.aboutMe && (
                <div className="profile-about">
                  <h3>About</h3>
                  <p>{profile.user.aboutMe}</p>
                </div>
              )}
              
              <div className="profile-stats">
                <div className="stat">
                  <span className="stat-count">{profile.postsCount}</span>
                  <span className="stat-label">Posts</span>
                </div>
                <div className="stat clickable" onClick={openFollowersModal}>
                  <span className="stat-count">{profile.followersCount}</span>
                  <span className="stat-label">Followers</span>
                </div>
                <div className="stat clickable" onClick={openFollowingModal}>
                  <span className="stat-count">{profile.followingCount}</span>
                  <span className="stat-label">Following</span>
                </div>
              </div>
            </div>
            
            {profile.isPrivate && !profile.isFollowing && !profile.isCurrentUser ? (
              <div className="card posts-card">
                <div className="private-account-message">
                  <div className="private-icon">🔒</div>
                  <h3>This account is private</h3>
                  <p>Follow @{profile.user.nickname || profile.user.firstName?.toLowerCase()} to see their posts and activity.</p>
                  <div className="private-stats">
                    <span>Only approved followers can see posts</span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="card posts-card">
                <h2 className="card-title">
                  {profile.isCurrentUser ? 'My Posts' : `${profile.user.firstName}'s Posts`}
                </h2>
                
                {profile.posts && profile.posts.length > 0 ? (
                  <div className="posts-container">
                    {profile.posts.map(post => (
                      <div key={post.id} className="post-card">
                        <div className="post-header">
                          <div className="post-author">
                            <div 
                              className="author-avatar clickable-avatar"
                              onClick={() => router.push(`/profile/${post.userId}`)}
                            >
                              {post.author.avatar ? (
                                <img src={mediaURL(post.author.avatar)} alt={`${post.author.firstName}'s avatar`} />
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
                        </div>
                        
                        <div className="post-content"><ContentActions kind="post" id={post.id} reportable={!profile.isCurrentUser} /><LikeButton id={post.id} count={post.likeCount} liked={post.isLiked} />
                          {post.content && <p className="post-text">{post.content}</p>}
                          {post.image && (
                            <div className="post-image-container">
                              <img
                                src={post.image.startsWith('http') ? post.image : `http://localhost:8080${encodeURI(post.image.replace(/\\/g, '/'))}`}
                                alt="Post image"
                                className="post-image"
                              />
                            </div>
                          )}
                        </div>
                        
                    
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="empty-posts-container">
                    {profile.user.isPrivate && !profile.isCurrentUser ? (
                      <div className="privacy-message">
                        <div className="privacy-icon">🔒</div>
                        <h4>Posts are private</h4>
                        <p>
                          {profile.isFollowing
                            ? "This user hasn't shared any posts with you yet."
                            : "Follow this account to see their posts."
                          }
                        </p>
                      </div>
                    ) : (
                      <p className="empty-posts">No posts yet.</p>
                    )}
                  </div>
                )}
              </div>
            )}

            {/* Add Followers Modal */}
            {activeModal === 'followers' && profile && (
              <div className="modal-overlay" onClick={closeModal}>
                <div className="modal-content" onClick={e => e.stopPropagation()}>
                  <div className="modal-header">
                    <h3>Followers</h3>
                    <button className="modal-close" onClick={closeModal}>×</button>
                  </div>
                  <div className="modal-body">
                    {loadingUserDetails ? (
                      <p className="loading-text">Loading followers...</p>
                    ) : profile.followers.length === 0 ? (
                      <p className="empty-list">No followers yet.</p>
                    ) : (
                      <ul className="follow-list">
                        {profile.followers.map((follower) => (
                          <li key={follower.id} className="follow-item" onClick={() => navigateToProfile(follower.id)}>
                            <div className="follow-avatar">
                              {follower.avatar ? (
                                <img src={mediaURL(follower.avatar)} alt={`${follower.firstName}'s avatar`} />
                              ) : (
                                <div className="avatar-placeholder">
                                  {(follower.firstName?.charAt(0).toUpperCase() ?? '?')}
                                </div>
                              )}
                            </div>
                            <div className="follow-info">
                              <p className="follow-name">
                                {follower.nickname || `${follower.firstName} ${follower.lastName}`}
                              </p>
                            </div>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                </div>
              </div>
            )}

            {/* Add Following Modal */}
            {activeModal === 'following' && profile && (
              <div className="modal-overlay" onClick={closeModal}>
                <div className="modal-content" onClick={e => e.stopPropagation()}>
                  <div className="modal-header">
                    <h3>Following</h3>
                    <button className="modal-close" onClick={closeModal}>×</button>
                  </div>
                  <div className="modal-body">
                    {loadingUserDetails ? (
                      <p className="loading-text">Loading following...</p>
                    ) : profile.following.length === 0 ? (
                      <p className="empty-list">Not following anyone yet.</p>
                    ) : (
                      <ul className="follow-list">
                        {profile.following.map((following) => (
                          <li key={following.id} className="follow-item" onClick={() => navigateToProfile(following.id)}>
                            <div className="follow-avatar">
                              {following.avatar ? (
                                <img src={mediaURL(following.avatar)} alt={`${following.firstName}'s avatar`} />
                              ) : (
                                <div className="avatar-placeholder">
                                  {(following.firstName?.charAt(0).toUpperCase() ?? '?')}
                                </div>
                              )}
                            </div>
                            <div className="follow-info">
                              <p className="follow-name">
                                {following.nickname || `${following.firstName} ${following.lastName}`}
                              </p>
                            </div>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  )
}
