#!/bin/bash

docker build -t cassandra-x .

docker stop cql
docker rm cql

docker run -d --rm -p 9042:9042 --name cql cassandra-x
