การทำงานจะเป็นดังนี้:

01. เมื่อมี client เชื่อมต่อเข้ามา:
    * ผ่านการ authenticate ด้วย token
    * สร้าง Client object ใหม่
    * ลงทะเบียนกับ Hub ผ่าน register channel

02. การรับส่งข้อความ:
    * Client.readPump() อ่านข้อความจาก WebSocket
    * ส่งไปยัง Hub.broadcast channel
    * Hub กระจายข้อความไปยัง clients ที่เกี่ยวข้อง
    * Client.writePump() ส่งข้อความออกไป

03. การจัดการ Error:
    * ใช้ defer เพื่อ cleanup resources
    * จัดการ connection timeout
    * บันทึก error ลงใน log
    * ส่ง error trace ไปยัง OpenTelemetry

04. Clean Shutdown:
    * ปิด channels ทั้งหมด
    * ยกเลิกการลงทะเบียน clients
    * ปิด connections ที่คงอยู่
    *รอให้ goroutines ทำงานเสร็จ


1. Client Connect:
   Browser -> HTTP -> HandleWebSocket()
                  -> Upgrade to WebSocket
                  -> Create Client
                  -> Register with Hub

2. Client Send Message:
   Browser -> WebSocket -> Client.readPump()
                       -> Parse Message
                       -> ChatUseCase.HandleMessage()
                       -> Save to DB
                       -> Hub.broadcast

3. Server Broadcast:
   Hub.broadcast -> Client.writePump()
                 -> WebSocket
                 -> Browser

4. Client Disconnect:
   Browser Close -> Client.readPump() ends
                 -> Hub.unregister
                 -> Clean up