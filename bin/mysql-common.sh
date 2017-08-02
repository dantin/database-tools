
OPTIONS=dt:i:
LONG_OPTIONS=debug,tier:,instance:

validate_tier() {
	case "$1" in
		prod)
			;;
		prod_slave)
			;;
		dev5)
			;;
		*)
			echo "Invalid tier parameter!"
			exit 2
			;;
	esac
}

validate_instance() {
	case "$1" in
		db)
			;;
		user_credit)
			;;
		order)
			;;
		order_bj)
			;;
		order_sh)
			;;
		order_gz)
			;;
		order_cd)
			;;
		order_gl)
			;;
		*)
			echo "Invalid instance parameter!"
			exit 2
			;;
	esac
}

print_usage() {
	echo
	echo "Usage:"
	echo
	echo "  $0 -t [prod|prod_slave|dev5] -i [db|order|order_bj|order_sh|order_gz|order_cd|order_gl|user_credit] -d \"sql statement\""
	echo
}

parse_arguments() {
	parsed_args=$(getopt -o $OPTIONS --long $LONG_OPTIONS -n "$0" -- "$@")
	
	if [ $? != 0 ]; then
		echo "Parsing arguments error!!"
		print_usage
		exit 1
	fi
	
	eval set -- "${parsed_args}"
	
	while true; do
		case "$1" in
			-d|--debug)
				echo "Debug: debug mode ON"
				set -x
				shift
				;;
			-i|--instance)
				echo "Instance: $2"
				validate_instance $2
				instance=$2
				shift 2
				;;
			-t|--tier)
				echo "Tier: $2"
				validate_tier $2
				tier=$2
				shift 2
				;;
			--)
				shift
				break;
				;;
			*)
				echo "Internal error!"
				exit 1
				;;
		esac
	done
	
	sql="$@"
	
	if [[ -z "$tier" ]]; then
		echo "empty tier!"
		print_usage
		exit 1
	fi
	
	if [[ -z "$instance" ]]; then
		echo "empty instance!"
		print_usage
		exit 1
	fi
}

