#!/bin/bash

# see https://www.ibm.com/developerworks/community/blogs/ff78a96f-bf23-457e-befa-77f266844cbb/entry/db2cli_validate_command_line_tool_for_validating_and_testing_cli_environment_and_configuration?lang=en
# see https://blogs.sas.com/content/sgf/2017/11/16/connecting-sas-db2-database-via-odbc-without-tears/

CLIDRIVER=${1:-/opt/db2/clidriver}

export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$CLIDRIVER/lib
export DB2CLIINIPATH=$CLIDRIVER/cfg
export DB2DSDRIVER_CFG_PATH=$CLIDRIVER/cfg

$CLIDRIVER/bin/db2cli validate -dsn SAMPLE -connect -user db2inst1 -passwd P4ssw0rd
