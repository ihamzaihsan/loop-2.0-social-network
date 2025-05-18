import React, { useState, useEffect } from 'react';
import NotificationItem from './NotificationItem';
import styles from './NotificationDrawer.module.css';
import { WebSocketClient } from '../webscoket/websocket';

interface Notification {
  id: number;
  type: string;
  content: string;
  status: string;
  created_at: string;
  from_user_id?: number;
  sender_name?: string;
  sender_avatar?: string;
  group_title?: string;
  actions?: string[];
  related_id?: number;
}

interface NotificationDrawerProps {
  onClose: () => void;
  onNotificationRead: () => void;
}

const NotificationDrawer: React.FC<NotificationDrawerProps> = ({ onClose, onNotificationRead }) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchNotifications();
    
    // Set up WebSocket listener for real-time updates
    const wsClient = WebSocketClient.getInstance();
    
    // Handle follow request accepted/rejected updates
    wsClient.addMessageHandler('follow_request_handled', (content) => {
      console.log('Follow request handled:', content);
      
      // Remove the handled follow request from the notifications list
      if (content.request_id) {
        setNotifications(prevNotifications => 
          prevNotifications.filter(notification => 
            !(notification.type === 'follow_request' && notification.related_id === content.request_id)
          )
        );
        
        // Update the notification count
        onNotificationRead();
      }
    });
    
    return () => {
      // Clean up by removing the handler when component unmounts
      const wsClient = WebSocketClient.getInstance();
      wsClient.messageHandlers.delete('follow_request_handled');
    };
  }, [onNotificationRead]);

  const fetchNotifications = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:8080/notifications', {
        method: 'GET',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch notifications');
      }

      const data = await response.json();
      if (data.success) {
        // Ensure notifications is always an array
        const notifs = Array.isArray(data.notifications) ? data.notifications : [];
        setNotifications(notifs);
      } else {
        throw new Error(data.message || 'Failed to fetch notifications');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
      console.error('Error fetching notifications:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleMarkAllRead = async () => {
    try {
      const response = await fetch('http://localhost:8080/notifications/read-all', {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        // Update UI to mark all notifications as read
        setNotifications(prevNotifications =>
          prevNotifications.map(notification => ({
            ...notification,
            status: 'read'
          }))
        );
        
        // Update notification count in parent component
        onNotificationRead();
      }
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error);
    }
  };

  const handleNotificationAction = async (notificationId: number, action: string) => {
    try {
      const response = await fetch(`http://localhost:8080/notifications/action?id=${notificationId}`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ action }),
      });

      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          // Remove the notification from the list or mark it as actioned
          setNotifications(prevNotifications =>
            prevNotifications.filter(n => n.id !== notificationId)
          );
          
          // Update notification count in parent component
          onNotificationRead();
        }
      }
    } catch (error) {
      console.error(`Failed to ${action} notification:`, error);
    }
  };

  const handleMarkAsRead = async (notificationId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/notifications/read?id=${notificationId}`, {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        // Update the status of the notification in the UI
        setNotifications(prevNotifications =>
          prevNotifications.map(notification =>
            notification.id === notificationId
              ? { ...notification, status: 'read' }
              : notification
          )
        );
        
        // Update notification count in parent component
        onNotificationRead();
      }
    } catch (error) {
      console.error('Failed to mark notification as read:', error);
    }
  };

  return (
    <div className={styles.drawer}>
      <div className={styles.header}>
        <h3>Notifications</h3>
        <div className={styles.actions}>
          <button
            className={styles.markAllButton}
            onClick={handleMarkAllRead}
            disabled={!notifications || notifications.filter(n => n.status === 'unread').length === 0}
          >
            Mark all as read
          </button>
          <button className={styles.closeButton} onClick={onClose}>
            &times;
          </button>
        </div>
      </div>

      <div className={styles.content}>
        {loading ? (
          <div className={styles.loading}>Loading notifications...</div>
        ) : error ? (
          <div className={styles.error}>{error}</div>
        ) : notifications.length === 0 ? (
          <div className={styles.empty}>No notifications yet</div>
        ) : (
          <ul className={styles.notificationList}>
            {notifications.map(notification => (
              <NotificationItem
                key={notification.id}
                notification={notification}
                onAction={handleNotificationAction}
                onMarkAsRead={handleMarkAsRead}
              />
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

export default NotificationDrawer; 