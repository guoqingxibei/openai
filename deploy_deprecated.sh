#!/bin/bash -e
cd ~/openai

echo "Fetching openai code..."
git fetch && git reset --hard origin/main

echo "Building openai..."
go build -o openai
chown -R openai:openai ../openai

echo "Restarting openai..."
systemctl restart openai
echo "Restarted openai successfully"
