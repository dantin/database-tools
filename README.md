# database-tools

Database Tools

### importer

导入Mock数据

	./bin/importer --config ./misc/importer_conf.toml

### bulk_ddl

批量DDL操作

	# 批量操作
	./bin/bulk_ddl --config ./misc/bulk_ddl_config.toml
	# 批量回退
	./bin/bulk_ddl -R --config ./misc/bulk_ddl_config.toml

### 依赖包

	glide install
	# 解析parser
	cd vendor/github.com/pingcap/tidb
	make parser
