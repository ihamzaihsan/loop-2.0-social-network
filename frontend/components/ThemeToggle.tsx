'use client'

import { useSyncExternalStore } from 'react'
import { usePathname } from 'next/navigation'
import styles from './ThemeToggle.module.css'

const preferenceKey = 'loop-theme'
type Theme = 'light' | 'dark'

function applyPreference() {
    let preference: string | null = null
    try { preference = localStorage.getItem(preferenceKey) } catch { /* Storage may be unavailable. */ }
    const theme = preference === 'dark' || preference === 'light'
        ? preference
        : window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    document.documentElement.dataset.theme = theme
}

function subscribe(onChange: () => void) {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const refresh = () => { applyPreference(); onChange() }
    const storage = (event: StorageEvent) => {
        if (event.key === preferenceKey || event.key === null) refresh()
    }
    window.addEventListener('loop-theme-change', onChange)
    window.addEventListener('storage', storage)
    media.addEventListener('change', refresh)
    refresh()
    return () => {
        window.removeEventListener('loop-theme-change', onChange)
        window.removeEventListener('storage', storage)
        media.removeEventListener('change', refresh)
    }
}

function getTheme(): Theme {
    return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'
}

export default function ThemeToggle({ authOnly = false }: { authOnly?: boolean }) {
    const pathname = usePathname()
    const theme = useSyncExternalStore(subscribe, getTheme, () => 'light' as Theme)
    if (authOnly && pathname !== '/login' && pathname !== '/register' && pathname !== '/forgot-password' && pathname !== '/reset-password' && pathname !== '/auth/google/complete') return null

    const toggle = () => {
        const nextTheme = theme === 'dark' ? 'light' : 'dark'
        try { localStorage.setItem(preferenceKey, nextTheme) } catch { /* Keep the current tab usable. */ }
        document.documentElement.dataset.theme = nextTheme
        window.dispatchEvent(new Event('loop-theme-change'))
    }

    return (
        <button type="button" onClick={toggle} aria-pressed={theme === 'dark'}
            className={authOnly ? styles.authToggle : 'sidebar-item'} aria-label="Dark mode">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                {theme === 'dark' ? <><circle cx="12" cy="12" r="4" /><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5" /></>
                    : <path d="M21 12.8A9 9 0 1 1 11.2 3 7 7 0 0 0 21 12.8Z" />}
            </svg>
            <span>{theme === 'dark' ? 'Light mode' : 'Dark mode'}</span>
        </button>
    )
}
