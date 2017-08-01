# database-tools

Database Tools

### importer

导入Mock数据

	./bin/importer --config ./misc/importer_conf.toml

### 依赖包

	glide install
	# 解析parser
	cd vendor/github.com/pingcap/tidb
	make parser
