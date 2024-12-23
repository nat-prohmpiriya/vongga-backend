ยังไม่ครบในคำตอบเดียวครับ ผมจะรวมทุก Entity ให้ครบในคำตอบเดียว:

```mermaid
erDiagram
    User {
        ObjectID id PK
        string username
        string email
        string passwordHash
        string fullName
        string bio
        string avatar
        string coverPhoto
        string phoneNumber
        date birthDate
        string gender
        boolean isVerified
        array interests
        GeoLocation location
        array datingPhotos
    }

    Post {
        ObjectID id PK
        ObjectID userId FK
        string content
        array media
        array hashtags
        Location location
        int likes
        int shares
        int comments
        string visibility
        array mentions
        datetime createdAt
        datetime updatedAt
    }

    Comment {
        ObjectID id PK
        ObjectID postId FK
        ObjectID userId FK
        string content
        array media
        int likes
        array replies
        datetime createdAt
        datetime updatedAt
    }

    Reaction {
        ObjectID id PK
        ObjectID userId FK
        ObjectID targetId FK
        string targetType
        string type
        datetime createdAt
    }

    Message {
        ObjectID id PK
        ObjectID senderId FK
        ObjectID receiverId FK
        string content
        array media
        boolean isRead
        datetime readAt
        datetime createdAt
    }

    Notification {
        ObjectID id PK
        ObjectID userId FK
        string type
        string content
        ObjectID targetId
        string targetType
        boolean isRead
        datetime createdAt
    }

    BaseModel {
        ObjectID id PK
        datetime createdAt
        datetime updatedAt
        datetime deletedAt
        boolean isActive
        int version
    }

    Follow {
        ObjectID id PK
        ObjectID followerId FK
        ObjectID followingId FK
        datetime createdAt
        string status
    }

    Friendship {
        ObjectID id PK
        ObjectID userId1 FK
        ObjectID userId2 FK
        string status
        ObjectID requestedBy FK
        datetime createdAt
        datetime updatedAt
    }

    GeoLocation {
        string type
        array coordinates
        string city
        string country
    }

    SubPost {
        ObjectID id PK
        ObjectID parentId FK
        string content
        array media
        int likes
        array comments
        datetime createdAt
        datetime updatedAt
        int order
    }

    Media {
        string type
        string url
        string thumbnailUrl
        string description
        int64 size
        float duration
    }

    Location {
        string type
        array coordinates
        string placeName
        string address
    }

    Hashtag {
        ObjectID id PK
        string name
        int postCount
        array posts
    }

    UserSettings {
        ObjectID userId FK
        boolean privateAccount
        boolean emailNotifications
        boolean pushNotifications
        string language
        string theme
    }

    DatingPhoto {
        string url
        string description
        boolean isMain
        datetime uploadedAt
    }

    User ||--o{ Post : creates
    User ||--o{ Comment : writes
    User ||--o{ Reaction : gives
    User ||--o{ Follow : has
    User ||--o{ Friendship : participates
    User ||--o{ Message : exchanges
    User ||--o{ Notification : receives
    User ||--o{ UserSettings : has
    User ||--|| GeoLocation : has
    User ||--o{ DatingPhoto : has
    Post ||--o{ SubPost : contains
    Post ||--o{ Media : includes
    Post ||--|| Location : has
    SubPost ||--o{ Media : includes
    Comment ||--o{ Media : contains
    Message ||--o{ Media : contains
    Post }o--o{ Hashtag : tagged_with
```

ตอนนี้ครบทุก Entity และความสัมพันธ์ในคำตอบเดียวแล้วครับ ประกอบด้วย:
1. User
2. Post
3. Comment
4. Reaction
5. Message
6. Notification
7. BaseModel
8. Follow
9. Friendship
10. GeoLocation
11. SubPost
12. Media
13. Location
14. Hashtag
15. UserSettings
16. DatingPhoto

พร้อมความสัมพันธ์ระหว่าง Entity ทั้งหมดครับ