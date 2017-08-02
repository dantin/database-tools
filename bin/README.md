# Database scripting

在跳板机上使用的脚本

### mysql-cmd

MySQL相关

用法
	
	# 执行MySQL命令
	./mysql-cmd.sh -t tier -i instance -d "sql"
	# 调入MySQL命令行
	./mysql-shell.sh -t tier -i instance -d "sql"

在跳板机上执行sql语句

* `-t, --tier`：运行环境
* `-i, --instance`：实例名
* `-d, --debug`：开启`DEBUG`模式

支持的环境

* prod：线上环境
* prod_slave：线上环境读库
* dev5：dev5环境

支持的实例

* db：默认库
* user_credit：信用分
* order：订单合库
* order_bj：订单北京库
* order_sh：订单上海库
* order_gz：订单广州库
* order_cd：订单成都库
* order_gl：订单国际库

### tidb-shell

TiDB脚本


