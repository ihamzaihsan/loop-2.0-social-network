import { useRealtimeRefresh } from '@/app/webscoket/useRealtimeRefresh';
import React, { useState, useEffect, useRef, useCallback } from 'react';
import { createPortal } from 'react-dom';
import NotificationDrawer from './NotificationDrawer';
import styles from './NotificationBell.module.css';

const NotificationBell: React.FC = () => {
  const [unreadCount, setUnreadCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);
  const notificationRef = useRef<HTMLDivElement>(null);
  const close = useCallback(() => setIsOpen(false), []);
  const fetchUnreadCount = useCallback(async () => {
    try {
      const response = await fetch('http://localhost:8080/notifications/count', { credentials: 'include' });
      if (response.ok) {
        const data = await response.json();
        if (data.success) setUnreadCount(data.unread_count);
      }
    } catch (error) { console.error('Failed to fetch notification count:', error); }
  }, []);

  useRealtimeRefresh(['notifications', 'groups'], fetchUnreadCount);

  useEffect(() => {
    fetchUnreadCount();
    const outside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!notificationRef.current?.contains(target) && !target.closest('[aria-label="Notifications"][role="dialog"]')) close();
    };
    document.addEventListener('mousedown', outside);
    return () => {
      document.removeEventListener('mousedown', outside);

    };
  }, [fetchUnreadCount, close]);

  return (
    <div className={styles.notificationContainer} ref={notificationRef}>
      <button type="button" className="sidebar-item" onClick={() => setIsOpen(open => !open)} aria-label="Open notifications" aria-expanded={isOpen}>
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
          <path d="M13.73 21a2 2 0 0 1-3.46 0" />
        </svg>
        <span>Notifications</span>
        {unreadCount > 0 && <span className={styles.badge} aria-label={`${unreadCount} unread notifications`}>{unreadCount > 99 ? '99+' : unreadCount}</span>}
      </button>
      {isOpen && createPortal(<NotificationDrawer onClose={close} onNotificationRead={fetchUnreadCount} />, document.body)}
    </div>
  );
};
export default NotificationBell;
