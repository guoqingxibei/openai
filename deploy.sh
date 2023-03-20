#!/bin/bash

cd ~/openai
echo "Fetching Go code..."
git fetch && git reset --hard origin/main
echo "Building Go program..."
go build -o openaiBin

# Stop the running Go program
echo "Stopping Go program..."
pkill -f openaiBin

# Wait for the program to stop
sleep 5

# Start the Go program
echo "Starting Go program..."
GO_ENV=prod nohup ./openaiBin > openai.log 2>&1 &

# Check if the program started successfully
if [ $? -eq 0 ]; then
  echo "Go program restarted successfully"
else
  echo "Failed to restart Go program"
fi