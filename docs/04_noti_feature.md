# Notification Features

## Overview
The notification system in Vongga Platform is designed to keep users informed about various interactions within the platform. The system covers social interactions, content engagement, and mentions across different features.

## Notification Types

### 1. Post Interactions
- **Mentions in Posts**
  - Trigger: When a user is mentioned using @username in a post
  - Message: "mentioned you in a post"
  - Additional: Also triggers when mentioned in post edits
  - Note: Users don't receive notifications for mentioning themselves

### 2. Comment Interactions
- **Post Comments**
  - Trigger: When someone comments on a user's post
  - Message: "commented on your post"
  - Note: Post owner doesn't receive notifications for their own comments

- **Comment Replies**
  - Trigger: When someone replies to a user's comment
  - Message: "replied to your comment"
  - Note: Original commenter doesn't receive notifications for their own replies

- **Comment Mentions**
  - Trigger: When a user is mentioned using @username in a comment
  - Message: "mentioned you in a comment"
  - Note: Users don't receive notifications for mentioning themselves

### 3. Reaction Interactions
- **Post Reactions**
  - Trigger: When someone reacts to a user's post
  - Message: "reacted to your post"
  - Note: Post owner doesn't receive notifications for their own reactions

- **Comment Reactions**
  - Trigger: When someone reacts to a user's comment
  - Message: "reacted to your comment"
  - Note: Comment owner doesn't receive notifications for their own reactions

### 4. Social Interactions
- **Follow**
  - Trigger: When someone follows a user
  - Message: "started following you"
  - Note: Users can't follow themselves

- **Friend Requests**
  - Trigger: When someone sends a friend request
  - Message: "sent you a friend request"
  - Note: System prevents duplicate friend requests

- **Friend Request Acceptance**
  - Trigger: When someone accepts a user's friend request
  - Message: "accepted your friend request"
  - Note: Both users become friends after acceptance

## Technical Implementation

### Notification Structure
Each notification contains:
- Recipient ID: User who will receive the notification
- Sender ID: User who triggered the notification
- Reference ID: ID of the related content (post, comment, etc.)
- Type: Type of notification (like, comment, follow, etc.)
- Reference Type: Context of the notification (post, comment)
- Message: Human-readable notification message
- Read Status: Whether the notification has been read
- Timestamp: When the notification was created

### Error Handling
- Failed notifications don't interrupt the main operation
- System logs notification failures for monitoring
- Users can still interact even if notification creation fails

### Performance Considerations
- Notifications are created asynchronously
- System includes pagination for notification retrieval
- Unread notifications count is cached for quick access

## API Endpoints

### Notification Management
```go
// NotificationUseCase interface
type NotificationUseCase interface {
    CreateNotification(recipientID, senderID, refID primitive.ObjectID, nType NotificationType, refType, message string) (*Notification, error)
    GetNotification(notificationID primitive.ObjectID) (*Notification, error)
    ListNotifications(recipientID primitive.ObjectID, limit, offset int) ([]Notification, error)
    MarkAsRead(notificationID primitive.ObjectID) error
    MarkAllAsRead(recipientID primitive.ObjectID) error
    DeleteNotification(notificationID primitive.ObjectID) error
    GetUnreadCount(recipientID primitive.ObjectID) (int64, error)
}
```

## Future Improvements
1. Real-time notifications using WebSocket
2. Notification preferences settings
3. Batch notification processing
4. Enhanced notification grouping
5. Rich media notifications
6. Custom notification templates
