# Story Feature Documentation

## Overview
Story feature คือฟีเจอร์ที่ให้ผู้ใช้สามารถแชร์รูปภาพหรือวิดีโอที่จะแสดงเป็นเวลา 24 ชั่วโมง คล้ายกับ Facebook Story โดยผู้ใช้สามารถดูได้ว่าใครเป็นคนดู story ของตัวเอง และ story จะถูกย้ายไป archive โดยอัตโนมัติหลังจากหมดอายุ

## Domain Model

### Story
```go
type Story struct {
    BaseModel    `bson:",inline"`
    UserID       string        `bson:"userId" json:"userId"`
    Media        StoryMedia    `bson:"media" json:"media"`
    Caption      string        `bson:"caption" json:"caption"`
    Location     string        `bson:"location" json:"location"`
    ViewersCount int          `bson:"viewersCount" json:"viewersCount"`
    Viewers      []StoryViewer `bson:"viewers" json:"viewers"`
    ExpiresAt    time.Time     `bson:"expiresAt" json:"expiresAt"`
    IsArchive    bool          `bson:"isArchive" json:"isArchive"`
    IsActive     bool          `bson:"isActive" json:"isActive"`
}
```

### StoryMedia
```go
type StoryMedia struct {
    URL       string    `bson:"url" json:"url"`
    Type      StoryType `bson:"type" json:"type"`
    Duration  int       `bson:"duration" json:"duration"`
    Thumbnail string    `bson:"thumbnail" json:"thumbnail"`
}
```

### StoryViewer
```go
type StoryViewer struct {
    UserID    string    `bson:"userId" json:"userId"`
    ViewedAt  time.Time `bson:"viewedAt" json:"viewedAt"`
    IsArchive bool      `bson:"isArchive" json:"isArchive"`
}
```

## Repository Layer
Story repository จัดการการเข้าถึงข้อมูลใน MongoDB โดยมีฟังก์ชันหลักดังนี้:
- `Create`: สร้าง story ใหม่
- `FindByID`: ค้นหา story ตาม ID
- `FindByUserID`: ค้นหา stories ทั้งหมดของผู้ใช้
- `FindActiveStories`: ค้นหา stories ที่ยังใช้งานได้
- `Update`: อัพเดทข้อมูล story
- `AddViewer`: เพิ่มผู้ชมและอัพเดทจำนวนผู้ชม
- `DeleteStory`: ลบ story (soft delete)
- `ArchiveExpiredStories`: เก็บ stories ที่หมดอายุเข้า archive

## Usecase Layer
Story usecase จัดการ business logic โดยมีฟังก์ชันหลักดังนี้:
- `CreateStory`: สร้าง story ใหม่พร้อมตรวจสอบความถูกต้องของข้อมูล
- `FindStoryByID`: ดึงข้อมูล story ตาม ID
- `FindUserStories`: ดึง stories ทั้งหมดของผู้ใช้
- `FindActiveStories`: ดึง stories ที่ยังใช้งานได้
- `ViewStory`: บันทึกการดู story พร้อมตรวจสอบว่าผู้ดูมีอยู่จริงและยังไม่เคยดู
- `DeleteStory`: ลบ story พร้อมตรวจสอบสิทธิ์การลบ
- `ArchiveExpiredStories`: เก็บ stories ที่หมดอายุเข้า archive

## HTTP Handler Layer
Story handler จัดการ HTTP requests โดยมี endpoints ดังนี้:

### Endpoints
1. `POST /api/stories`
   - สร้าง story ใหม่
   - Request Body:
     ```json
     {
       "mediaUrl": "string",
       "mediaType": "image|video",
       "mediaDuration": "int (optional)",
       "thumbnail": "string (optional)",
       "caption": "string (optional)",
       "location": "string (optional)"
     }
     ```

2. `GET /api/stories/active`
   - ดึง stories ที่ยังใช้งานได้ทั้งหมด

3. `GET /api/stories/user/:userId`
   - ดึง stories ของผู้ใช้ที่ระบุ

4. `GET /api/stories/:storyId`
   - ดึงข้อมูล story ตาม ID

5. `POST /api/stories/:storyId/view`
   - บันทึกการดู story

6. `DELETE /api/stories/:storyId`
   - ลบ story (เฉพาะเจ้าของ story)

## Security
- ทุก endpoint ต้องการ authentication
- มีการตรวจสอบสิทธิ์ในการลบ story
- มีการ validate ข้อมูลที่รับเข้ามาทุกครั้ง
- มีการตรวจสอบการหมดอายุของ story

## Business Rules
1. Story จะหมดอายุหลังจาก 24 ชั่วโมง
2. ผู้ใช้สามารถดู story ได้เพียงครั้งเดียว
3. มีเฉพาะเจ้าของ story ที่สามารถลบ story ได้
4. รองรับทั้งรูปภาพและวิดีโอ
5. Stories ที่หมดอายุจะถูกย้ายไป archive โดยอัตโนมัติ

## Future Improvements
1. เพิ่มการแจ้งเตือนเมื่อมีคนดู story
2. เพิ่มความสามารถในการตอบกลับ story
3. เพิ่มฟิลเตอร์และเอฟเฟกต์สำหรับรูปภาพและวิดีโอ
4. เพิ่มการแชร์ story ไปยังแพลตฟอร์มอื่น
5. เพิ่มการจัดกลุ่มผู้ชม story
