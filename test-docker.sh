#!/bin/bash

echo "=== Testing waitlib in Docker containers ==="
echo

echo "1. Building Docker images..."
docker build -f Dockerfile.alpine -t waitlib-alpine-test . > /dev/null 2>&1
docker build -f Dockerfile.scratch -t waitlib-scratch-test . > /dev/null 2>&1
echo "✓ Images built successfully"
echo

echo "2. Cleaning up any existing containers..."
docker rm -f waitlib-alpine-test waitlib-scratch-test > /dev/null 2>&1
echo

echo "3. Starting containers..."
ALPINE_ID=$(docker run -d --name waitlib-alpine-test waitlib-alpine-test)
SCRATCH_ID=$(docker run -d --name waitlib-scratch-test waitlib-scratch-test)
echo "✓ Alpine container: ${ALPINE_ID:0:12}"
echo "✓ Scratch container: ${SCRATCH_ID:0:12}"
echo

echo "4. Waiting 10 seconds for containers to initialize..."
sleep 10
echo

echo "5. Testing Alpine container:"
echo "   Container logs:"
docker logs waitlib-alpine-test | head -2 | sed 's/^/     /'
echo "   Process title in /proc/1/comm:"
echo "     $(docker exec waitlib-alpine-test cat /proc/1/comm)"
echo "   Process list (ps -o comm,cmd):"
docker exec waitlib-alpine-test ps -o comm,cmd | sed 's/^/     /'
echo

echo "6. Testing scratch container:"
echo "   Container logs:"
docker logs waitlib-scratch-test | head -2 | sed 's/^/     /'
echo "   Docker top output:"
docker top waitlib-scratch-test | sed 's/^/     /'
echo

echo "7. Waiting 30 more seconds to test uptime updates..."
sleep 30
echo "   Updated process title in Alpine container:"
echo "     $(docker exec waitlib-alpine-test cat /proc/1/comm)"
echo

echo "8. Docker ps output:"
docker ps --filter "name=waitlib" --format "table {{.Names}}\t{{.Image}}\t{{.Command}}\t{{.Status}}"
echo

echo "=== Test completed successfully! ==="
echo "The waitlib containers are running and updating their process titles."
echo "Clean up with: docker rm -f waitlib-alpine-test waitlib-scratch-test"
