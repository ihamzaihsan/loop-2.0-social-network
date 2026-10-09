'use client'
import { API } from "../utils/api";

import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'
import Link from 'next/link'
import NotificationBell from '../app/components/NotificationBell'
import BrandMark from './BrandMark'
import { WebSocketClient } from '../app/webscoket/websocket'
import { useRealtimeRefresh } from '../app/webscoket/useRealtimeRefresh'
import styles from './Sidebar.module.css'
import ThemeToggle from './ThemeToggle'

export default function Sidebar({ activePage }: { activePage: string }) {
    const router = useRouter()
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
    const [isMobile, setIsMobile] = useState(false)
    const [isConnected, setIsConnected] = useState(false)
    const [unreadMessages, setUnreadMessages] = useState(0)

    const refreshUnreadMessages = async () => {
        try {
            const response = await fetch(`${API}/chat/contacts`, { credentials: 'include' })
            if (!response.ok) return
            const data = await response.json()
            setUnreadMessages((data.contacts ?? []).reduce((total: number, contact: { unreadCount: number }) => total + (contact.unreadCount ?? 0), 0))
        } catch (error) { console.error('Failed to fetch unread messages:', error) }
    }
    useEffect(() => { refreshUnreadMessages() }, [])
    useRealtimeRefresh(['chat'], refreshUnreadMessages)

    useEffect(() => {
        const client = WebSocketClient.getInstance()
        client.connect()
        setIsConnected(client.socket?.readyState === WebSocket.OPEN)
        const updateConnection = (event: Event) => setIsConnected((event as CustomEvent).detail.connected)
        window.addEventListener('connection_status', updateConnection)
        return () => window.removeEventListener('connection_status', updateConnection)
    }, [])

    useEffect(() => {
        const checkScreenSize = () => {
            setIsMobile(window.innerWidth <= 820)
            if (window.innerWidth > 820) {
                setIsMobileMenuOpen(false)
            }
        }

        checkScreenSize()
        window.addEventListener('resize', checkScreenSize)
        return () => window.removeEventListener('resize', checkScreenSize)
    }, [])

    const handleLogout = async () => {
        try {
            const response = await fetch(`${API}/logout`, {
                method: 'POST',
                credentials: 'include'
            })

            if (response.ok) {
                localStorage.removeItem('sessionToken')
                WebSocketClient.resetInstance()
                router.push('/')
            } else {
                console.error('Error logging out')
            }
        } catch (error) {
            console.error('Logout failed:', error)
        }
    }

    const toggleMobileMenu = () => {
        setIsMobileMenuOpen(!isMobileMenuOpen)
    }

    const closeMobileMenu = () => {
        setIsMobileMenuOpen(false)
    }

    return (
        <>
            {/* Mobile hamburger button */}
            {isMobile && (
                <button
                    className="mobile-menu-toggle"
                    onClick={toggleMobileMenu}
                    aria-label="Toggle navigation menu"
                    aria-expanded={isMobileMenuOpen}
                    aria-controls="main-navigation"
                >
                    <div className={`hamburger ${isMobileMenuOpen ? 'active' : ''}`}>
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                    {!isMobileMenuOpen && unreadMessages > 0 && <span className={styles.mobileBadge} aria-label={`${unreadMessages} unread messages`}>{unreadMessages > 99 ? '99+' : unreadMessages}</span>}
                </button>
            )}

            {/* Mobile overlay */}
            {isMobile && isMobileMenuOpen && (
                <div className="mobile-overlay" onClick={closeMobileMenu}></div>
            )}

            <div id="main-navigation" inert={isMobile && !isMobileMenuOpen} className={`sidebar ${isMobile && isMobileMenuOpen ? 'mobile-open' : ''}`}>
                <div className="sidebar-logo">
                    <BrandMark inverse />
                    <p>Your social circle, in one place.</p>
                </div>
            <nav className="sidebar-nav">
                <Link href="/home" onClick={closeMobileMenu}>
                    <div className={`sidebar-item ${activePage === 'home' ? 'active' : ''}`}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
                            <polyline points="9 22 9 12 15 12 15 22"></polyline>
                        </svg>
                        <span>Home</span>
                    </div>
                </Link>

                <Link href="/profile" onClick={closeMobileMenu}>
                    <div className={`sidebar-item ${activePage === 'profile' ? 'active' : ''}`}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                            <circle cx="12" cy="7" r="4"></circle>
                        </svg>
                        <span>Profile</span>
                    </div>
                </Link>

                <Link href="/find-friends" onClick={closeMobileMenu}>
                    <div className={`sidebar-item ${activePage === 'friends' || activePage === 'find-friends' ? 'active' : ''}`}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                            <circle cx="9" cy="7" r="4"></circle>
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                        </svg>
                        <span>Friends</span>
                    </div>
                </Link>

                <Link href="/chat" onClick={closeMobileMenu}>
                    <div className={`sidebar-item ${activePage === 'chat' ? 'active' : ''}`}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                        </svg>
                        <span>Chat</span>
                        {unreadMessages > 0 && <span className={styles.badge} aria-label={`${unreadMessages} unread messages`}>{unreadMessages > 99 ? '99+' : unreadMessages}</span>}
                    </div>
                </Link>

                <Link href="/groups" onClick={closeMobileMenu}>
                    <div className={`sidebar-item ${activePage === 'groups' ? 'active' : ''}`}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M17 21v-2a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4v2"></path>
                            <circle cx="9" cy="7" r="4"></circle>
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                        </svg>
                        <span>Groups</span>
                    </div>
                </Link>

                <NotificationBell />
                <Link href="/search" onClick={closeMobileMenu}>
                    <div className={`sidebar-item ${activePage === 'search' ? 'active' : ''}`}>
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                            <circle cx="10" cy="10" r="7" />
                            <path d="m15 15 6 6" />
                        </svg>
                        <span>Search</span>
                    </div>
                </Link>

                <Link href="/settings" onClick={closeMobileMenu}>
                    <div className={`sidebar-item ${activePage === 'settings' ? 'active' : ''}`}>
                        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                            <path d="M4 7h16M4 17h16" />
                            <circle cx="8" cy="7" r="3" />
                            <circle cx="16" cy="17" r="3" />
                        </svg>
                        <span>Settings</span>
                    </div>
                </Link>
                <ThemeToggle />

                <div className="sidebar-footer">
                    <div className="sidebar-footer-note">
                        <span className={`status-dot ${isConnected ? '' : 'status-dot--offline'}`}></span>
                        <span role="status">{isConnected ? 'Connected to your circle' : 'Reconnecting to your circle…'}</span>
                    </div>
                    <button type="button" className="sidebar-item logout-item" onClick={() => { handleLogout(); closeMobileMenu(); }}>
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                            <polyline points="16 17 21 12 16 7"></polyline>
                            <line x1="21" y1="12" x2="9" y2="12"></line>
                        </svg>
                        <span>Logout</span>
                    </button>
                </div>
            </nav>
        </div>
        </>
    )
}
