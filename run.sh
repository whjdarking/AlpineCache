#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server -port=8000 &
./server -port=8001 &
./server -port=8002 -api=1 &

sleep 2
echo ">>> start getting from cache"
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &

sleep 2
echo ">>> retrieve once more "
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &
curl "http://localhost:9000/api?key=Tom" &
wait