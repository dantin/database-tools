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

expect -c "
	spawn mysql -h ${host} ${dbname} -u ${user} -P ${port} -p"$password";
	interact;
"

