#!/bin/bash

docker-compose -f ./docker-compose.dev.yml down 
# countedown 3 seconds
for i in $(seq 3 -1 1); do
    echo -n "$i ..."
    sleep 1
done
# run docker compose
docker-compose -f docker-compose.dev.yml up --build --force-recreate

# docker-compose -f ./docker-compose.dev.yml  down backend

# air -c ./.air.toml


#!/bin/bash

# สีสำหรับ output
# GREEN='\033[0;32m'
# RED='\033[0;31m'
# NC='\033[0m' # No Color

# echo -e "${GREEN}Starting debug process...${NC}"

# # 1. ตรวจสอบว่า container กำลังทำงานอยู่หรือไม่
# if ! docker ps | grep -q "vongga-backend-backend-1"; then
#     echo -e "${RED}Container is not running. Starting containers...${NC}"
#     docker-compose -f ./docker-compose.dev.yml up -d
#     sleep 3
# fi

# # 2. เช็คโครงสร้างไฟล์ใน container
# echo -e "${GREEN}Checking container file structure...${NC}"
# docker exec vongga-backend-backend-1 sh -c "
#     echo 'Working Directory:' && pwd && 
#     echo '\nRoot Directory Contents:' && ls -la /app &&
#     echo '\nEnvironment Variables:' && env &&
#     echo '\nChecking .env file:' && cat /app/.env
# "

# # 3. เปิด interactive shell ถ้าต้องการ debug เพิ่มเติม
# echo -e "${GREEN}\nOpening interactive shell...${NC}"
# docker exec -it vongga-backend-backend-1 sh
