#!/bin/bash

docker exec -it ignite \
  /opt/ignite/apache-ignite/bin/control.sh \
  --activate \
  --user ignite \
  --password ignite
