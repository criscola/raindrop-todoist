#!/bin/bash

DB_FILE=/root/db/mydb.sqlite3

# If database doesn't exist yet, load .sql file
if [ ! -e $DB_FILE ]; then
  sqlite3 $DB_FILE < /root/scripts/create_script.sql
fi

sqlite3