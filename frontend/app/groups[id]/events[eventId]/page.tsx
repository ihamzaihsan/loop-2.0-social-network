'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import styles from './eventDetail.module.css'
import Sidebar from '@/components/Sidebar'

interface EventDetail {
    id: number
    group_id: number
    title: string
    description: string
    event_time: string
    created_at: string
    user_response?: string
    response_options: EventResponseOption[]
}

interface EventResponseOption {
    id: number
    event_id: number
    option_text: string
    response_count: number
}

interface EventResponse {
    id: number
    user_id: number
    first_name: string
    last_name: string
    avatar?: string
    response: string
    created_at: string
}

export default function EventDetail({ params }: { params: { id: string, eventId: string } }) {
    const [event, setEvent] = useState<EventDetail | null>(null)
    const [responses, setResponses] = useState<EventResponse[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')

    const router = useRouter()
    const groupId = parseInt(params.id)
    const eventId = parseInt(params.eventId)

    useEffect(() => {
        const fetchEventDetails = async () => {
            try {
                const response = await fetch(`http://localhost:8080/groups/events/details?event_id=${eventId}`, {
                    credentials: 'include'
                })

                if (!response.ok) {
                    if (response.status === 404) {
                        router.push(`/groups/${groupId}`)
                        return
                    }
                    throw new Error('Failed to fetch event details')
                }

                const data = await response.json()
                if (data.success) {
                    setEvent(data.event)
                    setResponses(data.responses || [])
                }
            } catch (err) {
                setError('Failed to load event details')
                console.error(err)
            } finally {
                setLoading(false)
            }
        }

        fetchEventDetails()
    }, [eventId, groupId, router])

    const handleEventResponse = async (optionId: number) => {
        try {
            const response = await fetch('http://localhost:8080/groups/events/respond', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    event_id: eventId,
                    option_id: optionId
                })
            })

            if (!response.ok) {
                throw new Error('Failed to respond to event')
            }

            // Refresh event details to update response counts
            const updatedResponse = await fetch(`http://localhost:8080/groups/events/details?event_id=${eventId}`, {
                credentials: 'include'
            })
            const updatedData = await updatedResponse.json()
            if (updatedData.success) {
                setEvent(updatedData.event)
                setResponses(updatedData.responses || [])
            }
        } catch (err) {
            setError('Failed to respond to event')
            console.error(err)
        }
    }

    const formatDate = (dateString: string) => {
        const date = new Date(dateString)
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }

    if (loading) {
        return (
            <div className={styles.container}>
                <Sidebar activePage="groups" />
                <main className={styles.content}>
                    <div className={styles.loading}>Loading event details...</div>
                </main>
            </div>
        )
    }

    if (!event) {
        return (
            <div className={styles.container}>
                <Sidebar activePage="groups" />
                <main className={styles.content}>
                    <div className={styles.error}>Event not found or you don't have access.</div>
                    <button
                        className={styles.backButton}
                        onClick={() => router.push(`/groups/${groupId}`)}
                    >
                        Back to Group
                    </button>
                </main>
            </div>
        )
    }

    return (
        <div className={styles.container}>
            <Sidebar activePage="groups" />

            <main className={styles.content}>
                {error && <div className={styles.error}>{error}</div>}

                <div className={styles.header}>
                    <button
                        className={styles.backButton}
                        onClick={() => router.push(`/groups/${groupId}`)}
                    >
                        ← Back to Group
                    </button>
                </div>

                <div className={styles.eventCard}>
                    <h1 className={styles.eventTitle}>{event.title}</h1>

                    <div className={styles.eventMeta}>
                        <div className={styles.metaItem}>
                            <strong>When:</strong> {formatDate(event.event_time)}
                        </div>
                        <div className={styles.metaItem}>
                            <strong>Created:</strong> {formatDate(event.created_at)}
                        </div>
                    </div>

                    {event.description && (
                        <div className={styles.eventDescription}>
                            <p>{event.description}</p>
                        </div>
                    )}

                    <div className={styles.responseSection}>
                        <h2>Your Response</h2>

                        <div className={styles.responseOptions}>
                            {event.response_options.map(option => (
                                <button
                                    key={option.id}
                                    className={`${styles.responseOption} ${event.user_response === option.option_text ? styles.selectedResponse : ''}`}
                                    onClick={() => handleEventResponse(option.id)}
                                >
                                    {option.option_text}
                                </button>
                            ))}
                        </div>

                        {event.user_response && (
                            <p className={styles.currentResponse}>
                                You are currently marked as: <strong>{event.user_response}</strong>
                            </p>
                        )}
                    </div>

                    <div className={styles.attendeesSection}>
                        <h2>Responses</h2>

                        <div className={styles.responseStats}>
                            {event.response_options.map(option => (
                                <div key={option.id} className={styles.statItem}>
                                    <div className={styles.statLabel}>{option.option_text}</div>
                                    <div className={styles.statValue}>{option.response_count}</div>
                                </div>
                            ))}
                        </div>

                        <div className={styles.responseList}>
                            {responses.length > 0 ? (
                                responses.map(response => (
                                    <div key={response.id} className={styles.responseItem}>
                                        <div className={styles.userInfo}>
                                            {response.avatar ? (
                                                <img src={response.avatar} alt={`${response.first_name} ${response.last_name}`} className={styles.userAvatar} />
                                            ) : (
                                                <div className={styles.userInitials}>
                                                    {response.first_name[0]}{response.last_name[0]}
                                                </div>
                                            )}
                                            <div>
                                                <h3>{response.first_name} {response.last_name}</h3>
                                                <p className={styles.responseTime}>Responded: {formatDate(response.created_at)}</p>
                                            </div>
                                        </div>
                                        <div className={styles.userResponse}>
                                            <span className={styles.responseTag}>{response.response}</span>
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <div className={styles.emptyState}>
                                    <p>No responses yet.</p>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}
