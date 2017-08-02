#!/bin/bash

. ./tidb-common.sh

#############
#  Main
#############

tier=""
instance=""
sql=""

parse_arguments "$@"

. ./db/tidb/$tier/$instance.sh

mysql -h ${host} ${dbname} -u ${user} -p"$password" -P ${port} <<EOF
$sql
EOF

