# Vongga Backend

A robust authentication service built with Go, following Clean Architecture and Domain-Driven Design principles.

## Tech Stack

- **Language**: Go
- **Framework**: Fiber
- **Database**: MongoDB
- **Cache**: Redis
- **Authentication**: Firebase
- **Documentation**: Swagger UI

## Architecture

The project follows Clean Architecture with the following layers:

```
.
├── config/             # Configuration and initialization
├── delivery/           # HTTP handlers and middleware
│   └── http/
│       ├── handler/    # HTTP handlers
│       └── middleware/ # HTTP middleware
├── domain/            # Business logic interfaces and entities
├── repository/        # Data access layer
├── usecase/          # Business logic implementation
└── docs/             # Swagger documentation
```

## Features

- Firebase Authentication integration
- JWT-based authentication with refresh tokens
- MongoDB for user data storage
- Redis for refresh token management
- Swagger UI for API documentation
- CORS support
- Clean Architecture pattern
- Domain-Driven Design

## Authentication Flow

1. **Login Flow**:
   - Frontend authenticates with Firebase
   - Frontend sends Firebase ID token to backend
   - Backend verifies Firebase token
   - Backend creates/updates user in MongoDB
   - Backend generates JWT access & refresh tokens
   - Backend stores refresh token in Redis
   - Returns tokens to frontend

2. **Token Usage**:
   - Access Token:
     - Short-lived (1 hour)
     - Used for API authentication
     - Sent in Authorization header
   - Refresh Token:
     - Long-lived (30 days)
     - Stored in Redis
     - Used to get new access tokens
     - Can be revoked for logout

## API Endpoints

### Authentication

- **POST** `/api/auth/login`
  - Login with Firebase token
  - Returns user data and tokens

- **POST** `/api/auth/refresh`
  - Find new access token using refresh token
  - Returns new token pair

- **POST** `/api/auth/logout`
  - Revoke refresh token
  - Invalidates the session

### Users

- **GET** `/api/users/profile`
  - Find user profile
  - Requires authentication

- **POST** `/api/users`
  - Create or update user
  - Requires authentication

## Findting Started

1. **Prerequisites**:
   - Go 1.21 or higher
   - MongoDB
   - Redis
   - Firebase project credentials

2. **Environment Variables**:
   ```env
   SERVER_ADDRESS=:8080
   MONGO_URI=mongodb://localhost:27017
   MONGO_DB=vongga
   REDIS_URI=localhost:6379
   REDIS_PASSWORD=
   FIREBASE_CREDENTIALS_PATH=path/to/firebase-credentials.json
   JWT_SECRET=your-secret-key
   REFRESH_TOKEN_SECRET=your-refresh-secret-key
   ```

3. **Run the Application**:
   ```bash
   # Install dependencies
   go mod download

   # Run the server
   go run main.go
   ```

4. **API Documentation**:
   - Swagger UI available at: `http://localhost:8080/swagger/`

## Security Features

- Firebase token verification
- JWT-based authentication
- Refresh token rotation
- Token revocation support
- Secure password handling
- CORS configuration
- Rate limiting (TODO)

## Development

1. **Adding New Endpoints**:
   - Create handler in `delivery/http/handler/`
   - Add route in `main.go`
   - Add Swagger documentation
   - Implement business logic in usecase layer

2. **Database Changes**:
   - Update entity in `domain/`
   - Update repository interface
   - Implement changes in repository layer

3. **Generate Swagger Docs**:
   ```bash
   swag init
   ```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
