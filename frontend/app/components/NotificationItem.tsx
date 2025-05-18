import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { formatDistanceToNow } from 'date-fns';
import styles from './NotificationItem.module.css';

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
}

interface NotificationItemProps {
  notification: Notification;
  onAction: (notificationId: number, action: string) => void;
  onMarkAsRead: (notificationId: number) => void;
}

const NotificationItem: React.FC<NotificationItemProps> = ({
  notification,
  onAction,
  onMarkAsRead,
}) => {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // Handle navigation to different pages based on notification type
  const handleClick = () => {
    // Mark the notification as read first
    if (notification.status === 'unread') {
      onMarkAsRead(notification.id);
    }

    // Navigate to appropriate page based on notification type
    switch (notification.type) {
      case 'follow_request':
        router.push('/followers');
        break;
      case 'group_invitation':
      case 'group_join_request':
      case 'group_event':
        if (notification.from_user_id) {
          router.push(`/groups/${notification.from_user_id}`);
        }
        break;
      default:
        // For other notification types or if no specific navigation is needed
        break;
    }
  };

  const getIcon = () => {
    switch (notification.type) {
      case 'follow_request':
        return <span className={`${styles.icon} ${styles.followIcon}`}>👤</span>;
      case 'group_invitation':
        return <span className={`${styles.icon} ${styles.groupIcon}`}>👥</span>;
      case 'group_join_request':
        return <span className={`${styles.icon} ${styles.joinIcon}`}>🔑</span>;
      case 'group_event':
        return <span className={`${styles.icon} ${styles.eventIcon}`}>📅</span>;
      default:
        return <span className={styles.icon}>📣</span>;
    }
  };

  const handleAction = async (action: string) => {
    setIsLoading(true);
    await onAction(notification.id, action);
    setIsLoading(false);
  };

  const getActionsUI = () => {
    if (!notification.actions || notification.actions.length === 0) {
      return null;
    }

    return (
      <div className={styles.actions}>
        {notification.actions.includes('accept') && (
          <button
            className={`${styles.actionButton} ${styles.acceptButton}`}
            onClick={() => handleAction('accept')}
            disabled={isLoading}
          >
            Accept
          </button>
        )}
        {notification.actions.includes('reject') && (
          <button
            className={`${styles.actionButton} ${styles.rejectButton}`}
            onClick={() => handleAction('reject')}
            disabled={isLoading}
          >
            Reject
          </button>
        )}
      </div>
    );
  };

  // Safe time formatting with fallback
  const formattedTime = notification.created_at 
    ? formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })
    : 'recently';

  return (
    <li
      className={`${styles.item} ${
        notification.status === 'unread' ? styles.unread : ''
      }`}
    >
      <div className={styles.content} onClick={handleClick}>
        <div className={styles.icon}>{getIcon()}</div>
        <div className={styles.details}>
          <p className={styles.message}>{notification.content}</p>
          <span className={styles.time}>{formattedTime}</span>
        </div>
      </div>
      {getActionsUI()}
    </li>
  );
};

export default NotificationItem; 