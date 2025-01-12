# Follow and Friendship Features

This document describes the follow and friendship features in the Vongga Platform.

## Follow Feature

The follow feature allows users to follow other users without requiring mutual consent. This is similar to the following system in platforms like Twitter or Instagram.

### Follow Actions

#### 1. Follow a User
- Endpoint: `POST /api/v1/follow/:id`
- Description: Follow another user
- Authentication: Required
- Parameters:
  - `id`: ID of the user to follow
- Response:
  - Success (200): `{"message": "Successfully followed user"}`
  - Error (400): Invalid user ID
  - Error (500): Internal server error

#### 2. Unfollow a User
- Endpoint: `POST /api/v1/unfollow/:id`
- Description: Unfollow a previously followed user
- Authentication: Required
- Parameters:
  - `id`: ID of the user to unfollow
- Response:
  - Success (200): `{"message": "Successfully unfollowed user"}`
  - Error (400): Invalid user ID
  - Error (500): Internal server error

#### 3. Find Followers
- Endpoint: `GET /api/v1/followers/:id`
- Description: Find list of users who follow the specified user
- Authentication: Optional
- Parameters:
  - `id`: User ID
  - `limit`: Number of results per page (default: 10)
  - `offset`: Pagination offset (default: 0)
- Response:
  - Success (200): FindMany of followers
  - Error (400): Invalid user ID
  - Error (500): Internal server error

#### 4. Find Following
- Endpoint: `GET /api/v1/following/:id`
- Description: Find list of users that the specified user follows
- Authentication: Optional
- Parameters:
  - `id`: User ID
  - `limit`: Number of results per page (default: 10)
  - `offset`: Pagination offset (default: 0)
- Response:
  - Success (200): FindMany of following users
  - Error (400): Invalid user ID
  - Error (500): Internal server error

### Block Actions

#### 1. Block a User
- Endpoint: `POST /api/v1/block/:id`
- Description: Block a user
- Authentication: Required
- Parameters:
  - `id`: ID of the user to block
- Response:
  - Success (200): `{"message": "Successfully blocked user"}`
  - Error (400): Invalid user ID
  - Error (500): Internal server error

#### 2. Unblock a User
- Endpoint: `POST /api/v1/unblock/:id`
- Description: Unblock a previously blocked user
- Authentication: Required
- Parameters:
  - `id`: ID of the user to unblock
- Response:
  - Success (200): `{"message": "Successfully unblocked user"}`
  - Error (400): Invalid user ID
  - Error (500): Internal server error

## Friendship Feature

The friendship feature implements a bidirectional relationship between users, requiring mutual consent. This is similar to the friend system in platforms like Facebook.

### Friendship Actions

#### 1. Send Friend Request
- Endpoint: `POST /api/v1/friends/request/:id`
- Description: Send a friend request to another user
- Authentication: Required
- Parameters:
  - `id`: ID of the user to send friend request to
- Response:
  - Success (200): `{"message": "Friend request sent successfully"}`
  - Error (400): Invalid user ID
  - Error (500): Internal server error

#### 2. Accept Friend Request
- Endpoint: `POST /api/v1/friends/accept/:id`
- Description: Accept a pending friend request
- Authentication: Required
- Parameters:
  - `id`: ID of the user who sent the friend request
- Response:
  - Success (200): `{"message": "Friend request accepted"}`
  - Error (400): Invalid user ID or request not found
  - Error (500): Internal server error

#### 3. Reject Friend Request
- Endpoint: `POST /api/v1/friends/reject/:id`
- Description: Reject a pending friend request
- Authentication: Required
- Parameters:
  - `id`: ID of the user who sent the friend request
- Response:
  - Success (200): `{"message": "Friend request rejected"}`
  - Error (400): Invalid user ID or request not found
  - Error (500): Internal server error

#### 4. Cancel Friend Request
- Endpoint: `POST /api/v1/friends/cancel/:id`
- Description: Cancel a previously sent friend request
- Authentication: Required
- Parameters:
  - `id`: ID of the user the request was sent to
- Response:
  - Success (200): `{"message": "Friend request canceled"}`
  - Error (400): Invalid user ID or request not found
  - Error (500): Internal server error

#### 5. Unfriend
- Endpoint: `POST /api/v1/friends/unfriend/:id`
- Description: Remove a user from friends list
- Authentication: Required
- Parameters:
  - `id`: ID of the user to unfriend
- Response:
  - Success (200): `{"message": "Successfully unfriended"}`
  - Error (400): Invalid user ID or not friends
  - Error (500): Internal server error

#### 6. Find Friends
- Endpoint: `GET /api/v1/friends/:id`
- Description: Find list of user's friends
- Authentication: Optional
- Parameters:
  - `id`: User ID
  - `limit`: Number of results per page (default: 10)
  - `offset`: Pagination offset (default: 0)
- Response:
  - Success (200): FindMany of friends
  - Error (400): Invalid user ID
  - Error (500): Internal server error

#### 7. Find Friend Requests
- Endpoint: `GET /api/v1/friends/requests`
- Description: Find list of pending friend requests
- Authentication: Required
- Parameters:
  - `limit`: Number of results per page (default: 10)
  - `offset`: Pagination offset (default: 0)
- Response:
  - Success (200): FindMany of pending friend requests
  - Error (400): Invalid user ID
  - Error (500): Internal server error

## Error Handling

All endpoints follow a consistent error handling pattern:
- Invalid input parameters return 400 Bad Request
- Authentication errors return 401 Unauthorized
- Authorization errors return 403 Forbidden
- Server errors return 500 Internal Server Error

## Logging

All actions are logged using the following pattern:
1. Input parameters are logged at the start of each operation
2. Errors are logged before returning error responses
3. Success results are logged before returning success responses

This logging pattern helps with:
- Debugging issues
- Monitoring system behavior
- Tracking user interactions
- Auditing security-related actions
