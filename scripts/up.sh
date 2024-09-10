#!/bin/bash

# Go one level up to where docker-compose.yml is located
cd ..

# Build and run containers
docker-compose up --build -d

# Clean up unused images after exit
docker image prune -f

