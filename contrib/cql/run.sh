#!/bin/bash

docker build -t cassandra-x .

docker run -d --rm -p 9042:9042 cassandra-x
