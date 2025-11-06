#!/bin/bash

echo "Testing Forum API endpoints..."
echo "================================"

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s http://localhost:8080/api/health
echo -e "\n"

# Test registration
echo "2. Testing user registration..."
curl -s -X POST http://localhost:8080/api/register \
  -d "username=testuser&email=test@example.com&password=testpass" \
  -H "Content-Type: application/x-www-form-urlencoded"
echo -e "\n"

# Test login
echo "3. Testing user login..."
LOGIN_RESPONSE=$(curl -s -c cookies.txt -X POST http://localhost:8080/api/login \
  -d "email=test@example.com&password=testpass" \
  -H "Content-Type: application/x-www-form-urlencoded")
echo $LOGIN_RESPONSE
echo -e "\n"

# Test categories
echo "4. Testing categories endpoint..."
curl -s http://localhost:8080/api/categories
echo -e "\n"

# Test post creation (with session cookie)
echo "5. Testing post creation..."
curl -s -b cookies.txt -X POST http://localhost:8080/api/posts \
  -d "title=Test Post&content=This is a test post content&categories=Технологии,Общие" \
  -H "Content-Type: application/x-www-form-urlencoded"
echo -e "\n"

# Test getting posts
echo "6. Testing posts endpoint..."
curl -s http://localhost:8080/api/posts
echo -e "\n"

# Test logout
echo "7. Testing logout..."
curl -s -b cookies.txt -X POST http://localhost:8080/api/logout
echo -e "\n"

# Clean up
rm -f cookies.txt

echo "API testing completed!" 