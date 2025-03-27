'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import styles from './postDetail.module.css'
import Sidebar from '@/components/Sidebar'

interface GroupPost {
    id: number
    group_id: number
    user_id: number
    content: string
    image?: string
    created_at: string
    first_name: string
    last_name: string
    avatar?: string
}

interface GroupComment {
    id: number
    post_id: number
    user_id: number
    content: string
    created_at: string
    first_name: string
    last_name: string
    avatar?: string
}

export default function PostDetail({ params }: { params: { id: string, postId: string } }) {
    const [post, setPost] = useState<GroupPost | null>(null)
    const [comments, setComments] = useState<GroupComment[]>([])
    const [newComment, setNewComment] = useState('')
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')

    const router = useRouter()
    const groupId = parseInt(params.id)
    const postId = parseInt(params.postId)

    useEffect(() => {
        const fetchPostDetails = async () => {
            try {
                // First get the post details
                const postResponse = await fetch(`http://localhost:8080/groups/posts?group_id=${groupId}`, {
                    credentials: 'include'
                })

                if (!postResponse.ok) {
                    throw new Error('Failed to fetch post details')
                }

                const postData = await postResponse.json()
                if (postData.success) {
                    const foundPost = postData.posts.find((p: GroupPost) => p.id === postId)
                    if (foundPost) {
                        setPost(foundPost)
                    } else {
                        throw new Error('Post not found')
                    }
                }

                // Then get the comments
                const commentsResponse = await fetch(`http://localhost:8080/groups/posts/comments?post_id=${postId}`, {
                    credentials: 'include'
                })

                if (!commentsResponse.ok) {
                    throw new Error('Failed to fetch comments')
                }

                const commentsData = await commentsResponse.json()
                if (commentsData.success) {
                    setComments(commentsData.comments || [])
                }
            } catch (err) {
                setError('Failed to load post details')
                console.error(err)
            } finally {
                setLoading(false)
            }
        }

        fetchPostDetails()
    }, [groupId, postId])

    const handleCreateComment = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!newComment.trim()) {
            return
        }

        try {
            const response = await fetch('http://localhost:8080/groups/posts/comments/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    post_id: postId,
                    content: newComment
                })
            })

            if (!response.ok) {
                throw new Error('Failed to create comment')
            }

            const data = await response.json()
            if (data.success) {
                // Refresh comments
                const updatedResponse = await fetch(`http://localhost:8080/groups/posts/comments?post_id=${postId}`, {
                    credentials: 'include'
                })
                const updatedData = await updatedResponse.json()
                setComments(updatedData.comments || [])

                // Reset form
                setNewComment('')
            }
        } catch (err) {
            setError('Failed to create comment')
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
                    <div className={styles.loading}>Loading post details...</div>
                </main>
            </div>
        )
    }

    if (!post) {
        return (
            <div className={styles.container}>
                <Sidebar activePage="groups" />
                <main className={styles.content}>
                    <div className={styles.error}>Post not found or you don't have access.</div>
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

                <div className={styles.postCard}>
                    <div className={styles.postHeader}>
                        <div className={styles.postAuthor}>
                            {post.avatar ? (
                                <img src={post.avatar} alt={`${post.first_name} ${post.last_name}`} className={styles.authorAvatar} />
                            ) : (
                                <div className={styles.authorInitials}>
                                    {post.first_name[0]}{post.last_name[0]}
                                </div>
                            )}
                            <div>
                                <h3>{post.first_name} {post.last_name}</h3>
                                <p className={styles.postTime}>{formatDate(post.created_at)}</p>
                            </div>
                        </div>
                    </div>

                    <div className={styles.postContent}>
                        <p>{post.content}</p>
                        {post.image && (
                            <img src={post.image} alt="Post attachment" className={styles.postImage} />
                        )}
                    </div>
                </div>

                <div className={styles.commentsSection}>
                    <h2>Comments ({comments.length})</h2>

                    <div className={styles.createCommentForm}>
                        <form onSubmit={handleCreateComment}>
                            <textarea
                                placeholder="Write a comment..."
                                value={newComment}
                                onChange={(e) => setNewComment(e.target.value)}
                                className={styles.commentInput}
                            />
                            <button type="submit" className={styles.commentButton}>
                                Comment
                            </button>
                        </form>
                    </div>

                    {comments.length > 0 ? (
                        <div className={styles.commentsList}>
                            {comments.map(comment => (
                                <div key={comment.id} className={styles.commentCard}>
                                    <div className={styles.commentHeader}>
                                        <div className={styles.commentAuthor}>
                                            {comment.avatar ? (
                                                <img src={comment.avatar} alt={`${comment.first_name} ${comment.last_name}`} className={styles.authorAvatar} />
                                            ) : (
                                                <div className={styles.authorInitials}>
                                                    {comment.first_name[0]}{comment.last_name[0]}
                                                </div>
                                            )}
                                            <div>
                                                <h3>{comment.first_name} {comment.last_name}</h3>
                                                <p className={styles.commentTime}>{formatDate(comment.created_at)}</p>
                                            </div>
                                        </div>
                                    </div>

                                    <div className={styles.commentContent}>
                                        <p>{comment.content}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className={styles.emptyState}>
                            <p>No comments yet. Be the first to comment!</p>
                        </div>
                    )}
                </div>
            </main>
        </div>
    )
}
