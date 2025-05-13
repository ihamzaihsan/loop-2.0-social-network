'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import './find-friends.css'
import Sidebar from '../../components/Sidebar'

interface User {
  id: number
  firstName: string
  lastName: string
  email: string
  nickname?: string
  following?: boolean
  pendingFollow?: boolean
  isPrivate?: boolean
}

interface FollowRequest {
  id: number
  followerID: number
  followedID: number
  status: string
  username: string
  avatar?: string
}

interface Friend {
  id: number
  followerID: number
  followedID: number
  status: string
  username: string
  avatar?: string
}

export default function FindFriends() {
  const router = useRouter()
  const [users, setUsers] = useState<User[]>([])
  const [followRequests, setFollowRequests] = useState<FollowRequest[]>([])
  const [friends, setFriends] = useState<Friend[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentUserId, setCurrentUserId] = useState<number | null>(null)
  const [activeTab, setActiveTab] = useState<'users' | 'requests' | 'friends'>('users')

  useEffect(() => {
    const fetchData = async () => {
      try {
        // First, get the current user profile to know who we are
        const profileResponse = await fetch('http://localhost:8080/profile', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          }
        })

        if (!profileResponse.ok) {
          if (profileResponse.status === 401) {
            router.push('/')
            return
          }
          throw new Error('Failed to fetch your profile')
        }

        const profileData = await profileResponse.json()
        setCurrentUserId(profileData.user.id)

        // Fetch users

        await fetchUsers(profileData.user.id)

        // Fetch follow requests
        try {
          await fetchFollowRequests();
        } catch (fetchError) {
          console.error("Error in fetchFollowRequests:", fetchError);
        }
        
        // Fetch friends (mutual connections)
        try {
          await fetchFriends();
        } catch (fetchError) {
          console.error("Error in fetchFriends:", fetchError);
        }
      } catch (err: any) {
        console.error('Error fetching data:', err)
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [router])

  const fetchUsers = async (currentUserId: number) => {
    try {
      const usersResponse = await fetch('http://localhost:8080/users', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        }
      })

      if (!usersResponse.ok) {
        throw new Error('Failed to fetch users')
      }

      const responseData = await usersResponse.json()

      if (!responseData.success || !responseData.users) {
        throw new Error('Unexpected response format from server')
      }

      const userData = responseData.users

      // Filter out the current user
      const filteredUsers = userData.filter((user: User) => user.id !== currentUserId)
      setUsers(filteredUsers.map((user: User) => ({
        ...user,
        following: user.following || false
      })))
    } catch (err: any) {
      console.error('Error fetching users:', err)
      setError(err.message)
    }
  }

  const fetchFollowRequests = async () => {
    try {
      const response = await fetch('http://localhost:8080/follow-requests', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        }
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Error response:', errorText);
        throw new Error(`Failed to fetch follow requests: ${response.status} ${errorText}`);
      }

      const data = await response.json();

      // Handle both array and object with requests property
      if (Array.isArray(data)) {
        console.log('Setting follow requests from array');
        setFollowRequests(data);
      } else if (data && data.requests) {
        console.log('Setting follow requests from data.requests');
        setFollowRequests(data.requests);
      } else {
        setFollowRequests([]);
      }
    } catch (err: any) {
      console.error('Error in fetchFollowRequests:', err);
      // Don't break the whole page, just set empty requests
      setFollowRequests([]);
    }
  };

  const fetchFriends = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/friends', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        }
      });

      console.log('Response received:', response.status);

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Error response:', errorText);
        throw new Error(`Failed to fetch friends: ${response.status} ${errorText}`);
      }

      const data = await response.json();

      // Check the structure of the data
      if (data && data.friends) {
        setFriends(data.friends);
      } else if (Array.isArray(data)) {
        setFriends(data);
      } else {
        setFriends([]);
      }

      // Note: This will show the previous state due to React's asynchronous state updates
      // You'll need to check the next render to see the updated state
    } catch (err: any) {
      console.error('Error in fetchFriends:', err);
      // Don't break the whole page, just set empty friends
      setFriends([]);
    }
  };

  const handleFollowToggle = async (userId: number) => {
    try {
      const user = users.find(u => u.id === userId);
      const isFollowing = user?.following;

      console.log('Toggling follow for user:', userId, 'Currently following:', isFollowing);

      if (isFollowing) {
        // Unfollow logic
        console.log('Sending unfollow request for user:', userId);
        const response = await fetch('http://localhost:8080/unfollow', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ followed_id: userId })
        });

        const responseText = await response.text();
        console.log('Unfollow response:', response.status, responseText);

        if (!response.ok) {
          throw new Error(`Failed to unfollow user: ${response.status} ${responseText}`);
        }

        // Update the users list to reflect the change
        setUsers(prevUsers =>
          prevUsers.map(user =>
            user.id === userId
              ? { ...user, following: false }
              : user
          )
        );
        
        // If this was a friend, refresh the friends list
        if (friends.some(friend => friend.followedID === userId)) {
          fetchFriends();
        }
      } else {
        // Follow logic
        console.log('Sending follow request for user:', userId);
        const requestBody = JSON.stringify({ followed_id: userId });
        console.log('Request body:', requestBody);
        
        const response = await fetch('http://localhost:8080/follow', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          body: requestBody
        });

        const responseText = await response.text();
        console.log('Follow response:', response.status, responseText);

        if (!response.ok) {
          throw new Error(`Failed to follow user: ${response.status} ${responseText}`);
        }

        // Try to parse the response as JSON
        let data;
        try {
          data = JSON.parse(responseText);
        } catch (e) {
          console.error('Failed to parse response as JSON:', e);
          throw new Error('Invalid response format from server');
        }

        // Update the users list based on the follow status
        setUsers(prevUsers =>
          prevUsers.map(user =>
            user.id === userId
              ? {
                ...user,
                following: data.status === 'accept',
                pendingFollow: data.status === 'pending'
              }
              : user
          )
        );

        // Show appropriate message for pending requests
        if (data.status === 'pending') {
          alert('Follow request sent. Waiting for user approval.');
        } else if (data.status === 'accept') {
          // Check if this created a new friendship (mutual follow)
          fetchFriends();
        }
      }
    } catch (err: any) {
      console.error('Error toggling follow status:', err);
      setError(err.message);
    }
  }

  const handleFollowRequest = async (requestId: number, action: 'accept' | 'reject') => {
    try {
      console.log(`Sending ${action} request for request ID: ${requestId}`)

      const response = await fetch(`http://localhost:8080/follow-request?id=${requestId}`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ status: action })
      })

      console.log('Response status:', response.status)

      if (!response.ok) {
        const errorText = await response.text()
        console.error('Error response:', errorText)
        throw new Error(`Failed to ${action} follow request: ${response.status} ${errorText}`)
      }

      const data = await response.json()
      console.log('Response data:', data)

      // Remove the request from the list
      setFollowRequests(prevRequests =>
        prevRequests.filter(request => request.id !== requestId)
      )

      // If accepted, update the users list if the user is in it
      if (action === 'accept') {
        const request = followRequests.find(r => r.id === requestId)
        if (request) {
          setUsers(prevUsers =>
            prevUsers.map(user =>
              user.id === request.followerID
                ? { ...user, following: true, pendingFollow: false }
                : user
            )
          )
          
          // Refresh the friends list as this might have created a new friendship
          fetchFriends();
        }
      }

      // Show success message
      alert(`Follow request ${action === 'accept' ? 'accepted' : 'rejected'} successfully`)
    } catch (err: any) {
      console.error(`Error ${action}ing follow request:`, err)
      alert(`Error ${action}ing follow request: ${err.message}`)
    }
  }
  const navigateToProfile = (userId: number) => {
    router.push(`/profile/${userId}`)
  }

  if (loading) return <div className="find-friends-page">Loading...</div>
  if (error) return <div className="find-friends-page">Error: {error}</div>

  return (
    <div className="find-friends-page">
      <Sidebar activePage="find-friends" />

      <main className="users-container">
        <div className="tabs">
        <button
            className={`tab-button ${activeTab === 'friends' ? 'active' : ''}`}
            onClick={() => setActiveTab('friends')}
          >
            Friends {friends.length > 0 && <span className="friends-badge">{friends.length}</span>}
          </button>

          <button
            className={`tab-button ${activeTab === 'users' ? 'active' : ''}`}
            onClick={() => setActiveTab('users')}
          >
            Find Friends
          </button>
          <button
            className={`tab-button ${activeTab === 'requests' ? 'active' : ''}`}
            onClick={() => setActiveTab('requests')}
          >
            Follow Requests {followRequests.length > 0 && <span className="request-badge">{followRequests.length}</span>}
          </button>
        </div>

        {activeTab === 'users' && (
          <>
            <h1 className="page-title">Find Friends</h1>

            {users.length === 0 ? (
              <p className="no-users-message">No other users found.</p>
            ) : (
              <ul className="users-list">
                {users.map(user => (
                  <li key={user.id} className="user-card">
                    <div
                      className="user-info"
                      onClick={() => navigateToProfile(user.id)}
                    >
                      <div className="user-avatar">
                        {(user.nickname || user.firstName).charAt(0).toUpperCase()}
                      </div>
                      <div className="user-details">
                        <h3 className="user-name">
                          {user.nickname || `${user.firstName} ${user.lastName}`}
                        </h3>
                      </div>
                    </div>
                    <button
                      className={`follow-button ${user.following ? 'following' : ''} ${user.pendingFollow ? 'pending' : ''}`}
                      onClick={() => handleFollowToggle(user.id)}
                    >
                      {user.following ? 'Unfollow' : user.pendingFollow ? 'Pending' : 'Follow'}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </>
        )}

        {activeTab === 'requests' && (
          <>
            <h1 className="page-title">Follow Requests</h1>

            {followRequests.length === 0 ? (
              <p className="no-requests-message">No pending follow requests.</p>
            ) : (
              <ul className="requests-list">
                {followRequests.map(request => (
                  <li key={request.id} className="request-card">
                    <div className="user-info">
                      <div className="user-avatar">
                        {request.avatar ? (
                          <img src={request.avatar} alt={request.username} />
                        ) : (
                          request.username.charAt(0).toUpperCase()
                        )}
                      </div>
                      <div className="user-details">
                        <h3 className="user-name">{request.username}</h3>
                        <p className="request-text">wants to follow you</p>
                      </div>
                    </div>
                    <div className="request-actions">
                      <button
                        className="accept-button"
                        onClick={() => handleFollowRequest(request.id, 'accept')}
                      >
                        Accept
                      </button>
                      <button
                        className="reject-button"
                        onClick={() => handleFollowRequest(request.id, 'reject')}
                      >
                        Reject
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </>
        )}

        {activeTab === 'friends' && (
          <>
            <h1 className="page-title">Friends</h1>

            {friends.length === 0 ? (
              <p className="no-friends-message">You don't have any friends yet. When you and another user follow each other, they'll appear here.</p>
            ) : (
              <ul className="friends-list">
                {friends.map(friend => {
                  return (
                    <li key={friend.id} className="friend-card">
                      <div
                        className="user-info"
                        onClick={() => navigateToProfile(friend.followedID)}
                      >
                        <div className="user-avatar">
                          {friend.avatar ? (
                            <img src={friend.avatar} alt={friend.username} />
                          ) : (
                            friend.username.charAt(0).toUpperCase()
                          )}
                        </div>
                        <div className="user-details">
                          <h3 className="user-name">{friend.username}</h3>
                          <p className="friend-status">Friend</p>
                        </div>
                      </div>
                      <button
                        className="follow-button following"
                        onClick={() => handleFollowToggle(friend.followedID)}
                      >
                        Unfollow
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </>
        )}
      </main>
    </div>
  )
}