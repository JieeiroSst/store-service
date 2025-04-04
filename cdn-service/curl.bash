#!/bin/bash
# File Service API Tests
# This script contains curl commands to test all endpoints of the file service
# Set the base URL for your API
BASE_URL="http://localhost:8080"
FILE_ID=""  # Will be populated after upload

echo "======== File Service API Tests ========"

# 1. Test uploading an image file
echo -e "\n1. Testing image upload"
UPLOAD_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/files" \
  -H "Content-Type: multipart/form-data" \
  -F "content=@./test_image.jpg" \
  -F 'metadata={"filename":"test_image.jpg","mime_type":"image/jpeg","file_type":"FILE_TYPE_IMAGE","metadata":{"author":"Test User","description":"Test image for API testing"}}')

echo "Upload response:"
echo $UPLOAD_RESPONSE | jq .

# Extract file_id from response
FILE_ID=$(echo $UPLOAD_RESPONSE | jq -r '.file_id')
echo "Uploaded file ID: $FILE_ID"

# 2. Test retrieving file metadata
echo -e "\n2. Testing get file metadata"
curl -s -X GET "${BASE_URL}/v1/files/${FILE_ID}" | jq .

# 3. Test retrieving file content
echo -e "\n3. Testing get file content"
echo "Downloading file to test_download.jpg"
curl -s -X GET "${BASE_URL}/v1/files/${FILE_ID}/content" --output test_download.jpg
echo "File downloaded. Check test_download.jpg"

# 4. Test listing files with no filters
echo -e "\n4. Testing list files (no filters)"
curl -s -X GET "${BASE_URL}/v1/files" | jq .

# 5. Test listing files with file type filter
echo -e "\n5. Testing list files (filtered by type = IMAGE)"
curl -s -X GET "${BASE_URL}/v1/files?file_type=FILE_TYPE_IMAGE" | jq .

# 6. Test listing files with time range filter
echo -e "\n6. Testing list files (filtered by time range)"
# Get current time in RFC3339 format
NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
ONE_DAY_AGO=$(date -u -d "1 day ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-1d +"%Y-%m-%dT%H:%M:%SZ")
curl -s -X GET "${BASE_URL}/v1/files?created_after=${ONE_DAY_AGO}&created_before=${NOW}" | jq .

# 7. Test pagination
echo -e "\n7. Testing pagination"
curl -s -X GET "${BASE_URL}/v1/files?page_size=2" | jq .

# 8. Test uploading a video file
echo -e "\n8. Testing video upload"
VIDEO_RESPONSE=$(curl -s -X POST "${BASE_URL}/v1/files" \
  -H "Content-Type: multipart/form-data" \
  -F "content=@./test_video.mp4" \
  -F 'metadata={"filename":"test_video.mp4","mime_type":"video/mp4","file_type":"FILE_TYPE_VIDEO","metadata":{"author":"Test User","description":"Test video for API testing"}}')

echo "Video upload response:"
echo $VIDEO_RESPONSE | jq .

# Extract video file_id from response
VIDEO_ID=$(echo $VIDEO_RESPONSE | jq -r '.file_id')
echo "Uploaded video ID: $VIDEO_ID"

# 9. Test fetching video metadata
echo -e "\n9. Testing get video metadata"
curl -s -X GET "${BASE_URL}/v1/files/${VIDEO_ID}" | jq .

# 10. Test deleting a file
echo -e "\n10. Testing file deletion"
curl -s -X DELETE "${BASE_URL}/v1/files/${FILE_ID}" -v

# 11. Verify file is deleted by trying to fetch it
echo -e "\n11. Verifying file is deleted"
curl -s -X GET "${BASE_URL}/v1/files/${FILE_ID}" | jq .

echo -e "\n======== Tests Completed ========"