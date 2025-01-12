# Chat Feature Documentation

## Overview
The chat feature provides real-time messaging capabilities using both RESTful APIs and WebSocket connections. It supports private chats, group chats, file sharing, and message notifications.

## WebSocket Connection
Connect to the WebSocket server to receive real-time updates:
```javascript
const ws = new WebSocket('ws://your-backend/api/chat/ws')
```

### WebSocket Message Format
```typescript
interface WebSocketMessage {
  type: string      // Message type ('message', 'file', 'typing', etc.)
  roomId: string    // Chat room ID
  senderId: string  // Sender's user ID
  content: string   // Message content
  data?: any        // Additional data (optional)
  createdAt: Date   // Timestamp
}
```

## REST API Endpoints

### Room Operations

#### Create Private Chat
```http
POST /api/chat/rooms/private
Content-Type: application/json

{
  "userId": "string"  // ID of the user to chat with
}
```

#### Create Group Chat
```http
POST /api/chat/rooms/group
Content-Type: application/json

{
  "name": "string",
  "memberIds": ["string"]
}
```

#### Find User's Chat Rooms
```http
GET /api/chat/rooms
```

#### Add Member to Group
```http
POST /api/chat/rooms/:roomId/members
Content-Type: application/json

{
  "userId": "string"
}
```

#### Remove Member from Group
```http
DELETE /api/chat/rooms/:roomId/members/:userId
```

### Message Operations

#### Send Text Message
```http
POST /api/chat/messages
Content-Type: application/json

{
  "roomId": "string",
  "content": "string"
}
```

#### Send File Message
```http
POST /api/chat/messages/file
Content-Type: multipart/form-data

roomId: string
file: File
```

#### Find Chat Messages
```http
GET /api/chat/rooms/:roomId/messages?limit=20&offset=0
```

## Data Models

### ChatRoom
```typescript
interface ChatRoom {
  id: string
  name: string
  type: 'private' | 'group'
  members: string[]
  createdAt: Date
  updatedAt: Date
  isActive: boolean
}
```

### ChatMessage
```typescript
interface ChatMessage {
  id: string
  roomId: string
  senderId: string
  type: string
  content: string
  fileUrl?: string
  fileType?: string
  fileSize?: number
  readBy: string[]
  createdAt: Date
  updatedAt: Date
  isActive: boolean
}
```

## Features

### Real-time Features (WebSocket)
- Message delivery
- Typing indicators
- Online/Offline status
- Read receipts
- New message notifications

### REST API Features
- Create/manage chat rooms
- Member management
- File sharing
- Message history
- Room settings

## Usage Examples

### Connect to WebSocket
```typescript
const ws = new WebSocket('ws://your-backend/api/chat/ws')

ws.onopen = () => {
  console.log('Connected to chat')
}

ws.onmessage = (event) => {
  const message = JSON.parse(event.data)
  // Handle incoming message
}

ws.onclose = () => {
  console.log('Disconnected from chat')
}
```

### Send Message
```typescript
// Via WebSocket (for text messages)
ws.send(JSON.stringify({
  type: 'message',
  roomId: 'room-id',
  content: 'Hello!',
  createdAt: new Date()
}))

// Via REST API (for files)
const formData = new FormData()
formData.append('file', file)
formData.append('roomId', roomId)

await fetch('/api/chat/messages/file', {
  method: 'POST',
  body: formData
})
```

### Create Group Chat
```typescript
const response = await fetch('/api/chat/rooms/group', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: 'Group Name',
    memberIds: ['user1', 'user2', 'user3']
  })
})
```

## Error Handling

The API returns standard HTTP status codes:
- 200: Success
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 500: Internal Server Error

Error response format:
```json
{
  "error": "Error message"
}
```

## Best Practices

1. WebSocket Connection:
   - Implement reconnection logic
   - Handle connection errors
   - Clean up connection on component unmount

2. Message Handling:
   - Implement message queuing for offline scenarios
   - Handle message delivery status
   - Implement retry logic for failed messages

3. File Sharing:
   - Validate file types and sizes
   - Implement progress indicators
   - Handle upload errors

4. Performance:
   - Implement pagination for message history
   - Cache frequently accessed data
   - Optimize WebSocket message size
