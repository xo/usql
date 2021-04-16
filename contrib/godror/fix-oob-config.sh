#!/bin/bash

# adds DISABLE_OOB=on to user's .sqlnet.ora config
#
# See:
#   https://github.com/oracle/docker-images/issues/1352
#   https://franckpachot.medium.com/19c-instant-client-and-docker-1566630ab20e

echo "DISABLE_OOB=ON" >> $HOME/.sqlnet.ora
