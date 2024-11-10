#!/bin/bash
docker compose down
docker volume rm mezan-db-vol
docker compose up -d