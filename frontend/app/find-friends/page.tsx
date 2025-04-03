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

export default function FindFriends() {
  const router = useRouter()
  const [users, setUsers] = useState<User[]>([])
  const [followRequests, setFollowRequests] = useState<FollowRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentUserId, setCurrentUserId] = useState<number | null>(null)
  const [activeTab, setActiveTab] = useState<'users' | 'requests'>('users')

  useEffect(() => {
    const fetchData = async () => {
      try {
        // First, get the current user profile to know who we are
        console.log("Fetching profile...");
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
        console.log("Fetching users...");
        await fetchUsers(profileData.user.id)

        // Skip fetching follow requests for now
        // await fetchFollowRequests()
        console.log("About to fetch follow requests...");
        try {
          await fetchFollowRequests();
          console.log("Follow requests fetched successfully");
        } catch (fetchError) {
          console.error("Error in fetchFollowRequests:", fetchError);
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
    console.log("Inside fetchFollowRequests function");
    try {
      console.log('Making fetch request to /follow-requests...');
      const response = await fetch('http://localhost:8080/follow-requests', {
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
        throw new Error(`Failed to fetch follow requests: ${response.status} ${errorText}`);
      }

      const data = await response.json();
      console.log('Parsed response data:', data);

      // Handle both array and object with requests property
      if (Array.isArray(data)) {
        console.log('Setting follow requests from array');
        setFollowRequests(data);
      } else if (data && data.requests) {
        console.log('Setting follow requests from data.requests');
        setFollowRequests(data.requests);
      } else {
        console.log('No valid follow requests data found, setting empty array');
        setFollowRequests([]);
      }
    } catch (err: any) {
      console.error('Error in fetchFollowRequests:', err);
      // Don't break the whole page, just set empty requests
      setFollowRequests([]);
    }
  };


  const handleFollowToggle = async (userId: number) => {
    try {
      const user = users.find(u => u.id === userId)
      const isFollowing = user?.following

      if (isFollowing) {
        // Unfollow logic
        const response = await fetch('http://localhost:8080/unfollow', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ followed_id: userId })
        })

        if (!response.ok) {
          throw new Error('Failed to unfollow user')
        }

        // Update the users list to reflect the change
        setUsers(prevUsers =>
          prevUsers.map(user =>
            user.id === userId
              ? { ...user, following: false }
              : user
          )
        )
      } else {
        // Follow logic
        const response = await fetch('http://localhost:8080/follow', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ followed_id: userId })
        })

        if (!response.ok) {
          throw new Error('Failed to follow user')
        }

        // Get the response data to check the status
        const data = await response.json()

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
        )

        // Show appropriate message for pending requests
        if (data.status === 'pending') {
          alert('Follow request sent. Waiting for user approval.')
        }
      }
    } catch (err: any) {
      console.error('Error toggling follow status:', err)
      setError(err.message)
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

        {activeTab === 'users' ? (
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
        ) : (
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
      </main>
    </div>
  )
}