#!/bin/bash

rm -f example.csv

usql "adodb://Microsoft.ACE.OLEDB.12.0/?Extended+Properties=\"Text;HDR=NO;FMT=Delimited\"" \
  -c "create table example.csv(f1 text, f2 text, f3 text);" \
  -c "insert into example.csv(f1, f2, f3), values ('a', 'b', 'c');" \
  -c "select * from example.csv;"
