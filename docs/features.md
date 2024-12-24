# Vongga Platform Features Documentation

## Post Features

### Core Functionality
- **Create Post**
  - สร้างโพสต์ใหม่พร้อมรูปภาพ/วิดีโอ
  - รองรับการติด tags
  - รองรับการระบุตำแหน่ง (Location)
  - กำหนดการมองเห็น (Visibility)

- **Edit Post**
  - แก้ไขเนื้อหา, รูปภาพ, tags, location
  - เก็บประวัติการแก้ไข (EditHistory)
  - มีการระบุสถานะการแก้ไข (IsEdited)

- **Delete Post**
  - ลบโพสต์และข้อมูลที่เกี่ยวข้อง

- **View Post**
  - ดูโพสต์เดี่ยว
  - ดูรายการโพสต์ตาม userID
  - มีการนับจำนวนการดู (ViewCount)

### Additional Features
- **SubPosts**
  - รองรับการสร้างโพสต์ย่อย
  - จัดเรียงตามลำดับ (Order)
  - มีการนับ likes และ comments แยกต่างหาก

- **Media Support**
  - รองรับหลายประเภท (image, video, audio)
  - มี thumbnail สำหรับวิดีโอ
  - เก็บข้อมูลขนาดไฟล์และระยะเวลา (สำหรับวิดีโอ/เสียง)

- **Social Features**
  - นับจำนวนการแชร์ (ShareCount)
  - ระบบ tags
  - ระบบ location

## Comment Features

### Core Functionality
- **Create Comment**
  - แสดงความคิดเห็นในโพสต์
  - รองรับการแนบรูปภาพ/วิดีโอ
  - รองรับการตอบกลับความคิดเห็น (Reply)

- **Edit Comment**
  - แก้ไขเนื้อหาและรูปภาพ/วิดีโอ

- **Delete Comment**
  - ลบความคิดเห็น

- **View Comments**
  - ดูความคิดเห็นเดี่ยว
  - ดูรายการความคิดเห็นแบบแบ่งหน้า (Pagination)

### Additional Features
- **Nested Comments**
  - รองรับการตอบกลับความคิดเห็น
  - เชื่อมโยงกับความคิดเห็นต้นทาง (ReplyTo)

- **Reaction Integration**
  - นับจำนวน reactions แยกตามประเภท
  - แสดงสถิติ reactions ในแต่ละความคิดเห็น

## Reaction Features

### Core Functionality
- **Create/Update Reaction**
  - เพิ่มหรือเปลี่ยน reaction ในโพสต์/ความคิดเห็น
  - รองรับ dynamic reaction types

- **Delete Reaction**
  - ลบ reaction ออกจากโพสต์/ความคิดเห็น

- **View Reactions**
  - ดู reaction เดี่ยว
  - ดูรายการ reactions ทั้งหมด

### Additional Features
- **Flexible Reaction Types**
  - รองรับการเพิ่มประเภท reaction ใหม่
  - นับจำนวน reactions แยกตามประเภท

## Technical Features

### Data Structure
- Comments เป็น separate collection
  - รองรับการเติบโตของจำนวน comments
  - มี index ตาม postId เพื่อการค้นหาที่มีประสิทธิภาพ
  - Post เก็บเฉพาะจำนวน comments (CommentCount)

- SubPosts เป็น separate collection
  - แยกการจัดเก็บเพื่อรองรับการเติบโต
  - มี index ตาม parentId และ order
  - รองรับการจัดลำดับ SubPosts
  - Post เก็บเฉพาะจำนวน subposts (SubPostCount)

### Logging
- บันทึก input และ output ของทุก operation
- ติดตามการทำงานของระบบ

### Database
- ใช้ MongoDB เป็นฐานข้อมูล
- ออกแบบ schema ให้เหมาะกับการใช้งาน
