import React, { useState, useEffect, useRef } from 'react';
import NotificationDrawer from './NotificationDrawer';
import styles from './NotificationBell.module.css';

const NotificationBell: React.FC = () => {
  const [unreadCount, setUnreadCount] = useState<number>(0);
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const notificationRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Fetch initial unread count
    fetchUnreadCount();

    // Set up listener for new notifications
    const handleNewNotification = () => {
      fetchUnreadCount();
    };

    window.addEventListener('notification', handleNewNotification);

    // Check for updates every minute
    const interval = setInterval(fetchUnreadCount, 60000);

    // Handle clicks outside the notification drawer to close it
    const handleClickOutside = (event: MouseEvent) => {
      if (
        notificationRef.current &&
        !notificationRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);

    return () => {
      window.removeEventListener('notification', handleNewNotification);
      document.removeEventListener('mousedown', handleClickOutside);
      clearInterval(interval);
    };
  }, []);

  const fetchUnreadCount = async () => {
    try {
      const response = await fetch('http://localhost:8080/notifications/count', {
        method: 'GET',
        credentials: 'include',
      });

      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          setUnreadCount(data.unread_count);
        }
      }
    } catch (error) {
      console.error('Failed to fetch notification count:', error);
    }
  };

  const toggleNotifications = () => {
    setIsOpen(!isOpen);
  };

  return (
    <div className={styles.notificationContainer} ref={notificationRef} onClick={toggleNotifications}>
      {/* Bell Icon SVG (similar to other sidebar icons) */}
      <svg 
        xmlns="http://www.w3.org/2000/svg" 
        width="24" 
        height="24" 
        viewBox="0 0 24 24" 
        fill="none" 
        stroke="currentColor" 
        strokeWidth="2" 
        strokeLinecap="round" 
        strokeLinejoin="round"
        className={styles.bellIcon}
      >
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
        <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
      </svg>
      
      {unreadCount > 0 && (
        <span className={styles.badge}>{unreadCount > 99 ? '99+' : unreadCount}</span>
      )}
      
      {isOpen && (
        <NotificationDrawer 
          onClose={() => setIsOpen(false)}
          onNotificationRead={fetchUnreadCount}
        />
      )}
    </div>
  );
};

export default NotificationBell; 