#!/bin/bash
set -xe
set -o pipefail

CURRENT_DIR=$(cd $(dirname $0);pwd)
export MYSQL_PWD="isucari"
export LANG="C.UTF-8"
cd $CURRENT_DIR

cat 01_schema.sql 02_categories.sql initial.sql | mysql --defaults-file=/dev/null -u isucari isucari
