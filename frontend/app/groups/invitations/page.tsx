'use client'

import { useState, useEffect } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter } from 'next/navigation'
import styles from './invitations.module.css'
import '../groupsSidebar.css'

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

    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true)

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

        fetchData()
    }, [])

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
        <div className="groups-page">
            <Sidebar activePage="groups" />

            <div className="groups-container">
                <div className="groups-sidebar">
                    <div className="groups-sidebar-header">
                        <h2>Group Invitations</h2>
                        <button
                            className="back-button"
                            onClick={() => router.push('/groups')}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M19 12H5M12 19l-7-7 7-7" />
                            </svg>
                            Back
                        </button>
                    </div>

                    <div className="groups-list-container">
                        <div className="groups-nav">
                            <button
                                className="groups-nav-button active"
                                onClick={() => router.push('/groups/invitations')}
                            >
                                Invitations
                            </button>
                            <button
                                className="groups-nav-button"
                                onClick={() => router.push('/groups')}
                            >
                                My Groups
                            </button>
                        </div>

                        {loading ? (
                            <div className="loading-message">Loading...</div>
                        ) : invitations.length === 0 && joinRequests.length === 0 ? (
                            <div className="empty-list-message">
                                <p>No pending invitations or requests.</p>
                                <button
                                    className="discover-groups-button"
                                    onClick={() => router.push('/groups/discover')}
                                >
                                    Discover Groups
                                </button>
                            </div>
                        ) : (
                            <>
                                {invitations.length > 0 && (
                                    <div className="invitations-section">
                                        <h3 className="section-title">Group Invitations</h3>
                                        <ul className="groups-list">
                                            {invitations.map((invitation) => (
                                                <li key={invitation.id} className="group-item">
                                                    <div className="group-avatar">
                                                        <div className="avatar-placeholder">
                                                            {invitation.group_title.charAt(0)}
                                                        </div>
                                                    </div>
                                                    <div className="group-info">
                                                        <span className="group-name">{invitation.group_title}</span>
                                                        <p className="invitation-meta">
                                                            Invited by {invitation.inviter_name} • {new Date(invitation.created_at).toLocaleDateString()}
                                                        </p>
                                                    </div>
                                                    <div className="invitation-actions">
                                                        <button
                                                            className="accept-button"
                                                            onClick={() => handleInvitationAction(invitation.id, invitation.group_id, 'accept')}
                                                            disabled={processingAction[`invitation-${invitation.id}-accept`]}
                                                        >
                                                            {processingAction[`invitation-${invitation.id}-accept`] ? 'Accepting...' : 'Accept'}
                                                        </button>
                                                        <button
                                                            className="decline-button"
                                                            onClick={() => handleInvitationAction(invitation.id, invitation.group_id, 'reject')}
                                                            disabled={processingAction[`invitation-${invitation.id}-reject`]}
                                                        >
                                                            {processingAction[`invitation-${invitation.id}-reject`] ? 'Declining...' : 'Decline'}
                                                        </button>
                                                    </div>
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}

                                {joinRequests.length > 0 && (
                                    <div className="requests-section">
                                        <h3 className="section-title">Join Requests</h3>
                                        <ul className="groups-list">
                                            {joinRequests.map((request) => (
                                                <li key={request.id} className="group-item">
                                                    <div className="user-avatar">
                                                        {request.avatar ? (
                                                            <img src={request.avatar} alt={`${request.first_name}'s avatar`} />
                                                        ) : (
                                                            <div className="avatar-placeholder">
                                                                {request.first_name.charAt(0)}
                                                            </div>
                                                        )}
                                                    </div>
                                                    <div className="group-info">
                                                        <span className="user-name">{request.first_name} {request.last_name}</span>
                                                        <p className="request-meta">
                                                            Wants to join {request.group_title} • {new Date(request.created_at).toLocaleDateString()}
                                                        </p>
                                                    </div>
                                                    <div className="request-actions">
                                                        <button
                                                            className="accept-button"
                                                            onClick={() => handleJoinRequestAction(request.id, request.group_id, request.user_id, 'accept')}
                                                            disabled={processingAction[`request-${request.id}-accept`]}
                                                        >
                                                            {processingAction[`request-${request.id}-accept`] ? 'Approving...' : 'Approve'}
                                                        </button>
                                                        <button
                                                            className="decline-button"
                                                            onClick={() => handleJoinRequestAction(request.id, request.group_id, request.user_id, 'reject')}
                                                            disabled={processingAction[`request-${request.id}-reject`]}
                                                        >
                                                            {processingAction[`request-${request.id}-reject`] ? 'Rejecting...' : 'Reject'}
                                                        </button>
                                                    </div>
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}
                            </>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}

