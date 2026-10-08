'use client'

import { useRealtimeRefresh } from '@/app/webscoket/useRealtimeRefresh'

import { useState, useEffect } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter } from 'next/navigation'
import './invitations.css'

interface GroupInvitation {
    id: number
    group_id: number
    group_title: string
    inviter_id: number
    inviter_name: string
    created_at: string
    status: string
}

interface GroupJoinRequest {
    id: number
    group_id: number
    group_title: string
    user_id: number
    first_name: string
    last_name: string
    avatar?: string
    created_at: string
    status: string
}

export default function GroupInvitationsPage() {
    const [invitations, setInvitations] = useState<GroupInvitation[]>([])
    const [joinRequests, setJoinRequests] = useState<GroupJoinRequest[]>([])
    const [loading, setLoading] = useState(true)
    const [processingAction, setProcessingAction] = useState<{ [key: string]: boolean }>({})
    const [error, setError] = useState('')
    const [success, setSuccess] = useState('')
    const router = useRouter()

    const fetchData = async () => {
        try {

            // Fetch invitations
            const invitationsResponse = await fetch('http://localhost:8080/groups/invitations', {
                method: 'GET',
                credentials: 'include'
            })

            // Fetch join requests for groups where user is creator
            const joinRequestsResponse = await fetch('http://localhost:8080/groups/join/requests', {
                method: 'GET',
                credentials: 'include'
            })

            if (invitationsResponse.ok) {
                const data = await invitationsResponse.json()
                if (data && data.invitations) {
                    setInvitations(data.invitations)
                } else {
                    setInvitations([])
                }
            } else {
                console.error('Failed to fetch invitations')
                setError('Failed to load invitations')
            }

            if (joinRequestsResponse.ok) {
                const data = await joinRequestsResponse.json()
                if (data && data.requests) {
                    setJoinRequests(data.requests)
                } else {
                    setJoinRequests([])
                }
            } else {
                console.error('Failed to fetch join requests')
            }
        } catch (error) {
            console.error('Error fetching data:', error)
            setError('Failed to load invitations and requests')
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => { fetchData() }, [])
    useRealtimeRefresh(['groups', 'profiles'], fetchData)


    const handleInvitationAction = async (invitationId: number, groupId: number, action: string) => {
        const actionKey = `invitation-${invitationId}-${action}`
        setProcessingAction(prev => ({ ...prev, [actionKey]: true }))
        setError('')
        setSuccess('')

        try {
            const response = await fetch('http://localhost:8080/groups/membership/handle', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    group_id: groupId,
                    action: action,
                    request_type: 'invitation'
                }),
            })

            if (response.ok) {
                setSuccess(`Successfully ${action === 'accept' ? 'accepted' : 'rejected'} the invitation`)
                setInvitations(prev => prev.filter(inv => inv.id !== invitationId))

                if (action === 'accept') {
                    // Refresh the groups list to show the newly joined group
                    setTimeout(() => {
                        router.push(`/groups/${groupId}`)
                    }, 1500)
                }
            } else {
                const errorData = await response.json()
                setError(errorData.error || `Failed to ${action} the invitation`)
            }
        } catch (err) {
            console.error(`Error ${action}ing invitation:`, err)
            setError(`An error occurred while ${action}ing the invitation`)
        } finally {
            setProcessingAction(prev => ({ ...prev, [actionKey]: false }))
        }
    }

    const handleJoinRequestAction = async (requestId: number, groupId: number, userId: number, action: string) => {
        const actionKey = `request-${requestId}-${action}`
        setProcessingAction(prev => ({ ...prev, [actionKey]: true }))
        setError('')
        setSuccess('')

        try {
            const response = await fetch('http://localhost:8080/groups/membership/handle', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    group_id: groupId,
                    user_id: userId,
                    action: action,
                    request_type: 'request'
                }),
            })

            if (response.ok) {
                setSuccess(`Successfully ${action === 'accept' ? 'approved' : 'rejected'} the join request`)
                setJoinRequests(prev => prev.filter(req => req.id !== requestId))
            } else {
                const errorData = await response.json()
                setError(errorData.error || `Failed to ${action} the join request`)
            }
        } catch (err) {
            console.error(`Error ${action}ing join request:`, err)
            setError(`An error occurred while ${action}ing the join request`)
        } finally {
            setProcessingAction(prev => ({ ...prev, [actionKey]: false }))
        }
    }

    return (
        <div className="home-page">
            <Sidebar activePage="groups" />

            <main className="main-content">
                <div className="dashboard">
                    <div className="card feed-card">
                        <div className="invitations-hero-section">
                            <div className="hero-background">
                                <div className="hero-pattern"></div>
                                <div className="hero-gradient"></div>
                            </div>
                            <div className="hero-content">
                                <div className="hero-text">
                                    <h1 className="hero-title">
                                        <span className="title-icon">📬</span>
                                        Group Invitations
                                    </h1>
                                    <p className="hero-description">
                                        Review and manage your pending group invitations and join requests. Connect with communities that interest you!
                                    </p>
                                    <div className="hero-stats">
                                        <div className="stat-item">
                                            <span className="stat-number">{invitations.length}</span>
                                            <span className="stat-label">Pending Invites</span>
                                        </div>
                                        <div className="stat-divider"></div>
                                        <div className="stat-item">
                                            <span className="stat-number">{joinRequests.length}</span>
                                            <span className="stat-label">Join Requests</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="hero-actions">
                                    <button
                                        className="hero-primary-button"
                                        onClick={() => router.push('/groups/discover')}
                                    >
                                        <div className="button-icon">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                <circle cx="11" cy="11" r="8"></circle>
                                                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                                            </svg>
                                        </div>
                                        <span>Discover Groups</span>
                                    </button>
                                    <button
                                        className="hero-secondary-button"
                                        onClick={() => router.push('/groups')}
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                            <path d="M19 12H5M12 19l-7-7 7-7" />
                                        </svg>
                                        Back to Groups
                                    </button>
                                </div>
                            </div>
                        </div>

                        <div className="invitations-content-section">
                            {error && (
                                <div className="error-message-enhanced">
                                    <div className="error-icon">⚠️</div>
                                    <div className="error-content">
                                        <h4>Error</h4>
                                        <p>{error}</p>
                                    </div>
                                </div>
                            )}

                            {success && (
                                <div className="success-message-enhanced">
                                    <div className="success-icon">✅</div>
                                    <div className="success-content">
                                        <h4>Success</h4>
                                        <p>{success}</p>
                                    </div>
                                </div>
                            )}

                            {loading ? (
                                <div className="loading-state">
                                    <div className="loading-spinner">
                                        <div className="spinner"></div>
                                    </div>
                                    <h3>Loading invitations...</h3>
                                    <p>Gathering your pending invitations</p>
                                </div>
                            ) : invitations.length === 0 && joinRequests.length === 0 ? (
                                <div className="empty-state-enhanced">
                                    <div className="empty-illustration">
                                        <div className="empty-icon">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                                                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                                                <polyline points="22,6 12,13 2,6"></polyline>
                                            </svg>
                                        </div>
                                        <div className="empty-sparkles">
                                            <span className="sparkle sparkle-1">✨</span>
                                            <span className="sparkle sparkle-2">💌</span>
                                            <span className="sparkle sparkle-3">🎉</span>
                                        </div>
                                    </div>
                                    <div className="empty-content">
                                        <h3>No Invitations Yet!</h3>
                                        <p>You're all caught up! No pending invitations or join requests at the moment. Explore groups to find communities you'd love to join.</p>
                                        <div className="empty-actions">
                                            <button className="empty-primary-button" onClick={() => router.push('/groups/discover')}>
                                                <span className="button-icon">🔍</span>
                                                Discover Amazing Groups
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <div className="invitations-showcase">
                                    {invitations.length > 0 && (
                                        <div className="invitations-section">
                                            <div className="section-header">
                                                <h2 className="section-title">
                                                    <span className="section-icon">📨</span>
                                                    Group Invitations
                                                </h2>
                                                <div className="section-badge">
                                                    {invitations.length} pending
                                                </div>
                                            </div>
                                            <div className="invitations-grid">
                                                {invitations.map((invitation, index) => (
                                                    <div
                                                        key={invitation.id}
                                                        className="invitation-card-enhanced"
                                                        style={{
                                                            animationDelay: `${index * 0.1}s`
                                                        }}
                                                    >
                                                        <div className="card-background">
                                                            <div className="card-pattern"></div>
                                                            <div className="card-glow"></div>
                                                        </div>
                                                        <div className="card-content">
                                                            <div className="invitation-header">
                                                                <div className="group-avatar-enhanced">
                                                                    <div className="avatar-background"></div>
                                                                    <div className="avatar-letter">
                                                                        {invitation.group_title.charAt(0)}
                                                                    </div>
                                                                    <div className="avatar-ring"></div>
                                                                </div>
                                                                <div className="invitation-badge">
                                                                    <span className="badge-icon">👋</span>
                                                                    <span className="badge-text">Invited</span>
                                                                </div>
                                                            </div>
                                                            <div className="invitation-details">
                                                                <h3 className="group-title-enhanced">{invitation.group_title}</h3>
                                                                <p className="invitation-meta-enhanced">
                                                                    <span className="inviter">Invited by {invitation.inviter_name}</span>
                                                                    <span className="date">{new Date(invitation.created_at).toLocaleDateString()}</span>
                                                                </p>
                                                            </div>
                                                            <div className="invitation-actions-enhanced">
                                                                <button
                                                                    className="accept-button-enhanced"
                                                                    onClick={() => handleInvitationAction(invitation.id, invitation.group_id, 'accept')}
                                                                    disabled={processingAction[`invitation-${invitation.id}-accept`]}
                                                                >
                                                                    {processingAction[`invitation-${invitation.id}-accept`] ? (
                                                                        <>
                                                                            <div className="button-spinner"></div>
                                                                            <span>Accepting...</span>
                                                                        </>
                                                                    ) : (
                                                                        <>
                                                                            <span className="button-icon">✅</span>
                                                                            <span>Accept</span>
                                                                        </>
                                                                    )}
                                                                </button>
                                                                <button
                                                                    className="decline-button-enhanced"
                                                                    onClick={() => handleInvitationAction(invitation.id, invitation.group_id, 'reject')}
                                                                    disabled={processingAction[`invitation-${invitation.id}-reject`]}
                                                                >
                                                                    {processingAction[`invitation-${invitation.id}-reject`] ? (
                                                                        <>
                                                                            <div className="button-spinner"></div>
                                                                            <span>Declining...</span>
                                                                        </>
                                                                    ) : (
                                                                        <>
                                                                            <span className="button-icon">❌</span>
                                                                            <span>Decline</span>
                                                                        </>
                                                                    )}
                                                                </button>
                                                            </div>
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}

                                    {joinRequests.length > 0 && (
                                        <div className="requests-section">
                                            <div className="section-header">
                                                <h2 className="section-title">
                                                    <span className="section-icon">🙋</span>
                                                    Join Requests
                                                </h2>
                                                <div className="section-badge">
                                                    {joinRequests.length} pending
                                                </div>
                                            </div>
                                            <div className="requests-grid">
                                                {joinRequests.map((request, index) => (
                                                    <div
                                                        key={request.id}
                                                        className="request-card-enhanced"
                                                        style={{
                                                            animationDelay: `${(invitations.length + index) * 0.1}s`
                                                        }}
                                                    >
                                                        <div className="card-background">
                                                            <div className="card-pattern"></div>
                                                            <div className="card-glow"></div>
                                                        </div>
                                                        <div className="card-content">
                                                            <div className="request-header">
                                                                <div className="user-avatar-enhanced">
                                                                    {request.avatar ? (
                                                                        <img src={request.avatar} alt={`${request.first_name}'s avatar`} />
                                                                    ) : (
                                                                        <div className="avatar-placeholder-enhanced">
                                                                            {request.first_name.charAt(0)}
                                                                        </div>
                                                                    )}
                                                                    <div className="avatar-ring"></div>
                                                                </div>
                                                                <div className="request-badge">
                                                                    <span className="badge-icon">🚪</span>
                                                                    <span className="badge-text">Wants to Join</span>
                                                                </div>
                                                            </div>
                                                            <div className="request-details">
                                                                <h3 className="user-name-enhanced">{request.first_name} {request.last_name}</h3>
                                                                <p className="request-meta-enhanced">
                                                                    <span className="group-name">Wants to join {request.group_title}</span>
                                                                    <span className="date">{new Date(request.created_at).toLocaleDateString()}</span>
                                                                </p>
                                                            </div>
                                                            <div className="request-actions-enhanced">
                                                                <button
                                                                    className="approve-button-enhanced"
                                                                    onClick={() => handleJoinRequestAction(request.id, request.group_id, request.user_id, 'accept')}
                                                                    disabled={processingAction[`request-${request.id}-accept`]}
                                                                >
                                                                    {processingAction[`request-${request.id}-accept`] ? (
                                                                        <>
                                                                            <div className="button-spinner"></div>
                                                                            <span>Approving...</span>
                                                                        </>
                                                                    ) : (
                                                                        <>
                                                                            <span className="button-icon">✅</span>
                                                                            <span>Approve</span>
                                                                        </>
                                                                    )}
                                                                </button>
                                                                <button
                                                                    className="reject-button-enhanced"
                                                                    onClick={() => handleJoinRequestAction(request.id, request.group_id, request.user_id, 'reject')}
                                                                    disabled={processingAction[`request-${request.id}-reject`]}
                                                                >
                                                                    {processingAction[`request-${request.id}-reject`] ? (
                                                                        <>
                                                                            <div className="button-spinner"></div>
                                                                            <span>Rejecting...</span>
                                                                        </>
                                                                    ) : (
                                                                        <>
                                                                            <span className="button-icon">❌</span>
                                                                            <span>Reject</span>
                                                                        </>
                                                                    )}
                                                                </button>
                                                            </div>
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}

