#!/bin/bash

# 测试 Auth 和 User 服务的登录和获取用户信息功能
# 使用颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API 地址
API_BASE="http://localhost:28256/api/v1"

# 测试用户信息
USERNAME="admin"
PASSWORD="admin123"

echo -e "${YELLOW}=== 测试 Auth & User 服务 ===${NC}\n"

# 1. 测试 Ping
echo -e "${YELLOW}[1] 测试 Ping 端点...${NC}"
PING_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE/ping")
HTTP_CODE=$(echo "$PING_RESPONSE" | tail -n 1)
BODY=$(echo "$PING_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Ping 成功${NC}"
    echo "响应: $BODY"
else
    echo -e "${RED}✗ Ping 失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应: $BODY"
fi
echo ""

# 2. 测试登录
echo -e "${YELLOW}[2] 测试登录...${NC}"
echo "用户名: $USERNAME"
echo "密码: $PASSWORD"

LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n 1)
BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 登录成功${NC}"
    echo "响应: $BODY"
    
    # 提取 access_token
    ACCESS_TOKEN=$(echo "$BODY" | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')
    REFRESH_TOKEN=$(echo "$BODY" | grep -o '"refresh_token":"[^"]*' | sed 's/"refresh_token":"//')
    
    if [ -z "$ACCESS_TOKEN" ]; then
        echo -e "${RED}✗ 未能提取 access_token${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Access Token: ${ACCESS_TOKEN:0:50}...${NC}"
    echo -e "${GREEN}Refresh Token: ${REFRESH_TOKEN:0:50}...${NC}"
else
    echo -e "${RED}✗ 登录失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应: $BODY"
    exit 1
fi
echo ""

# 3. 测试 /me 端点 (获取当前令牌信息)
echo -e "${YELLOW}[3] 测试 /me 端点...${NC}"
ME_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

HTTP_CODE=$(echo "$ME_RESPONSE" | tail -n 1)
BODY=$(echo "$ME_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ /me 请求成功${NC}"
    echo "响应: $BODY"
else
    echo -e "${RED}✗ /me 请求失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应: $BODY"
fi
echo ""

# 4. 测试获取用户信息
echo -e "${YELLOW}[4] 测试获取用户信息...${NC}"
USERINFO_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE/user/info" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

HTTP_CODE=$(echo "$USERINFO_RESPONSE" | tail -n 1)
BODY=$(echo "$USERINFO_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 获取用户信息成功${NC}"
    echo "响应: $BODY"
else
    echo -e "${RED}✗ 获取用户信息失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应: $BODY"
fi
echo ""

# 5. 测试刷新令牌
echo -e "${YELLOW}[5] 测试刷新令牌...${NC}"
REFRESH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")

HTTP_CODE=$(echo "$REFRESH_RESPONSE" | tail -n 1)
BODY=$(echo "$REFRESH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 刷新令牌成功${NC}"
    echo "响应: $BODY"
    
    # 提取新的 access_token
    NEW_ACCESS_TOKEN=$(echo "$BODY" | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')
    echo -e "${GREEN}新 Access Token: ${NEW_ACCESS_TOKEN:0:50}...${NC}"
else
    echo -e "${RED}✗ 刷新令牌失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应: $BODY"
fi
echo ""

# 6. 测试登出
echo -e "${YELLOW}[6] 测试登出...${NC}"
LOGOUT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/logout" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")

HTTP_CODE=$(echo "$LOGOUT_RESPONSE" | tail -n 1)
BODY=$(echo "$LOGOUT_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 登出成功${NC}"
    echo "响应: $BODY"
else
    echo -e "${RED}✗ 登出失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应: $BODY"
fi
echo ""

echo -e "${YELLOW}=== 测试完成 ===${NC}"

