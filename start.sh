#!/bin/bash
set -e

go build -o server cmd/server/main.go

echo " Lancement du serveur..."
./server
