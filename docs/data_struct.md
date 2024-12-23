ผมจะช่วยแปลง Data Structure เป็น Mermaid ER Diagram ครับ:

```mermaid
erDiagram
    User ||--o{ Post : creates
    User ||--o{ Comment : writes
    User ||--o{ Reaction : gives
    User ||--o{ Follow : has
    User ||--o{ Friendship : participates
    User ||--o{ Message : exchanges
    User ||--o{ Notification : receives
    User ||--o{ UserSettings : has

    Post ||--o{ Comment : contains
    Post ||--o{ Reaction : receives
    Post ||--o{ SubPost : has
    Post ||--o{ Media : includes
    Post }|--|| Location : has
    Post }o--o{ Hashtag : tagged_with

    Comment ||--o{ Reaction : receives
    Comment ||--o{ Media : contains
    Comment ||--o{ Comment : replies_to

    USER {
        ObjectID id PK
        string username
        string email
        string password
        string firstName
        string lastName
        string avatar
        string bio
        datetime createdAt
        datetime updatedAt
        int followersCount
        int followingCount
        int friendsCount
        string provider
        boolean emailVerified
        datetime dateOfBirth
        string gender
        array interestedIn
        object location
        string relationStatus
        float height
        array interests
        string occupation
        string education
        array photos
        boolean isVerified
        string status
        string phoneNumber
    }

    POST {
        ObjectID id PK
        ObjectID userId FK
        string content
        array media
        map reactionCounts
        array comments
        array subPosts
        datetime createdAt
        datetime updatedAt
        array tags
        object location
        string visibility
        int shareCount
        int viewCount
        boolean isEdited
        array editHistory
    }

    COMMENT {
        ObjectID id PK
        ObjectID userId FK
        ObjectID postId FK
        string content
        array media
        map reactionCounts
        datetime createdAt
        datetime updatedAt
        ObjectID replyTo FK
    }

    REACTION {
        ObjectID id PK
        ObjectID postId FK
        ObjectID commentId FK
        ObjectID userId FK
        string type
        datetime createdAt
        datetime updatedAt
    }

    MESSAGE {
        ObjectID id PK
        ObjectID senderId FK
        ObjectID receiverId FK
        string content
        array media
        datetime createdAt
        datetime readAt
    }

    NOTIFICATION {
        ObjectID id PK
        ObjectID userId FK
        string type
        ObjectID actorId FK
        ObjectID postId FK
        string content
        boolean isRead
        datetime createdAt
    }
```

ผมได้แสดงความสัมพันธ์หลักๆ ระหว่าง Entity ต่างๆ:

1. User เป็นศูนย์กลางที่เชื่อมโยงกับ Entity อื่นๆ
2. Post มีความสัมพันธ์กับ Comment, Reaction และ Media
3. Comment สามารถมี Reply (Self-referential)
4. แสดง Field หลักๆ ของแต่ละ Entity
5. ใช้สัญลักษณ์แสดงความสัมพันธ์:
   - ||--o{ : One-to-Many
   - }|--|| : One-to-One
   - }o--o{ : Many-to-Many

Diagram นี้จะช่วยให้เห็นภาพรวมของระบบและความสัมพันธ์ระหว่าง Entity ต่างๆ ได้ชัดเจนขึ้นครับ