#!/bin/bash

HOST="localhost"
PORT=6379

redis-cli -h ${HOST} -p ${PORT} SET "user_id:generator" 10000

for i in {1..10};
do
    uid=$(redis-cli -h ${HOST} -p ${PORT} INCR "user_id:generator")
    redis-cli -h ${HOST} -p ${PORT} HMSET "user:"${uid}":info" "name" ${uid} "password" ${uid}
done;

