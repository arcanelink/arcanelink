#!/bin/bash

# 文件上传功能测试脚本

BASE_URL="http://localhost:8080/_api/v1"
TOKEN=""

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "文件上传功能测试"
echo "=========================================="

# 1. 注册用户
echo -e "\n${YELLOW}1. 注册测试用户...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }')

echo "$REGISTER_RESPONSE" | jq .

TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}注册失败，尝试登录...${NC}"

  # 尝试登录
  LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
      "username": "testuser",
      "password": "testpass123"
    }')

  TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')

  if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}登录失败，退出测试${NC}"
    exit 1
  fi
fi

echo -e "${GREEN}✓ 认证成功${NC}"
echo "Token: $TOKEN"

# 2. 创建测试文件
echo -e "\n${YELLOW}2. 创建测试文件...${NC}"
TEST_FILE="/tmp/test_upload.txt"
echo "This is a test file for upload functionality." > "$TEST_FILE"
echo "Created at: $(date)" >> "$TEST_FILE"
echo -e "${GREEN}✓ 测试文件已创建: $TEST_FILE${NC}"

# 3. 上传文件
echo -e "\n${YELLOW}3. 上传文件...${NC}"
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/files" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$TEST_FILE")

echo "$UPLOAD_RESPONSE" | jq .

FILE_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.file_id')
FILE_URL=$(echo "$UPLOAD_RESPONSE" | jq -r '.url')

if [ "$FILE_ID" == "null" ] || [ -z "$FILE_ID" ]; then
  echo -e "${RED}✗ 文件上传失败${NC}"
  exit 1
fi

echo -e "${GREEN}✓ 文件上传成功${NC}"
echo "File ID: $FILE_ID"
echo "File URL: $FILE_URL"

# 4. 获取文件信息
echo -e "\n${YELLOW}4. 获取文件信息...${NC}"
FILE_INFO=$(curl -s -X GET "$BASE_URL/files/$FILE_ID/info" \
  -H "Authorization: Bearer $TOKEN")

echo "$FILE_INFO" | jq .

if [ "$(echo "$FILE_INFO" | jq -r '.file_id')" == "$FILE_ID" ]; then
  echo -e "${GREEN}✓ 文件信息获取成功${NC}"
else
  echo -e "${RED}✗ 文件信息获取失败${NC}"
fi

# 5. 下载文件
echo -e "\n${YELLOW}5. 下载文件...${NC}"
DOWNLOAD_FILE="/tmp/test_download.txt"
curl -s -X GET "$BASE_URL/files/$FILE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -o "$DOWNLOAD_FILE"

if [ -f "$DOWNLOAD_FILE" ]; then
  echo -e "${GREEN}✓ 文件下载成功${NC}"
  echo "下载的文件内容:"
  cat "$DOWNLOAD_FILE"

  # 验证文件内容
  if diff "$TEST_FILE" "$DOWNLOAD_FILE" > /dev/null; then
    echo -e "${GREEN}✓ 文件内容验证成功${NC}"
  else
    echo -e "${RED}✗ 文件内容不匹配${NC}"
  fi
else
  echo -e "${RED}✗ 文件下载失败${NC}"
fi

# 6. 注册第二个用户用于测试消息发送
echo -e "\n${YELLOW}6. 注册第二个测试用户...${NC}"
REGISTER2_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "password": "testpass123"
  }')

echo "$REGISTER2_RESPONSE" | jq .

# 7. 发送文件消息（私聊）
echo -e "\n${YELLOW}7. 发送文件消息（私聊）...${NC}"
MESSAGE_RESPONSE=$(curl -s -X POST "$BASE_URL/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"recipient\": \"@testuser2:localhost\",
    \"content\": {
      \"msgtype\": \"m.file\",
      \"body\": \"test_upload.txt\",
      \"url\": \"$FILE_URL\",
      \"info\": {
        \"size\": $(stat -f%z "$TEST_FILE" 2>/dev/null || stat -c%s "$TEST_FILE"),
        \"mimetype\": \"text/plain\"
      }
    }
  }")

echo "$MESSAGE_RESPONSE" | jq .

if [ "$(echo "$MESSAGE_RESPONSE" | jq -r '.msg_id')" != "null" ]; then
  echo -e "${GREEN}✓ 文件消息发送成功${NC}"
else
  echo -e "${RED}✗ 文件消息发送失败${NC}"
fi

# 8. 创建房间并发送文件消息
echo -e "\n${YELLOW}8. 创建房间...${NC}"
ROOM_RESPONSE=$(curl -s -X POST "$BASE_URL/rooms" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Room",
    "invite": ["@testuser2:localhost"]
  }')

echo "$ROOM_RESPONSE" | jq .

ROOM_ID=$(echo "$ROOM_RESPONSE" | jq -r '.room_id')

if [ "$ROOM_ID" != "null" ] && [ -n "$ROOM_ID" ]; then
  echo -e "${GREEN}✓ 房间创建成功${NC}"
  echo "Room ID: $ROOM_ID"

  # 9. 在房间中发送文件消息
  echo -e "\n${YELLOW}9. 在房间中发送文件消息...${NC}"
  ROOM_MESSAGE_RESPONSE=$(curl -s -X POST "$BASE_URL/rooms/$ROOM_ID/messages" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"content\": {
        \"msgtype\": \"m.file\",
        \"body\": \"test_upload.txt\",
        \"url\": \"$FILE_URL\",
        \"info\": {
          \"size\": $(stat -f%z "$TEST_FILE" 2>/dev/null || stat -c%s "$TEST_FILE"),
          \"mimetype\": \"text/plain\"
        }
      }
    }")

  echo "$ROOM_MESSAGE_RESPONSE" | jq .

  if [ "$(echo "$ROOM_MESSAGE_RESPONSE" | jq -r '.event_id')" != "null" ]; then
    echo -e "${GREEN}✓ 房间文件消息发送成功${NC}"
  else
    echo -e "${RED}✗ 房间文件消息发送失败${NC}"
  fi
else
  echo -e "${RED}✗ 房间创建失败${NC}"
fi

# 清理
echo -e "\n${YELLOW}清理测试文件...${NC}"
rm -f "$TEST_FILE" "$DOWNLOAD_FILE"
echo -e "${GREEN}✓ 清理完成${NC}"

echo -e "\n=========================================="
echo -e "${GREEN}测试完成！${NC}"
echo "=========================================="
