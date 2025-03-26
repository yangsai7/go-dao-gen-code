# go-dao-code-gen

根据数据库表生成dao层代码

## 依赖 ##

    $ go get golang.org/x/tools/cmd/goimports

## 使用方法 ##

`go-dao-code-gen` 是一个命令行工具，首先需要通过 `go get` 安装。

    go get -u gitlab.nolibox.com/skyteam/go-dao-code-gen

使用时可以参考命令帮助信息。
``` bash
    $ go-dao-code-gen -help

Usage 1:
	go-dao-code-gen -h 127.0.0.1 -u root -p 123456 -D dbname -o ./dao

	or

	go-dao-code-gen -dsn='root:123456@tcp(127.0.0.1:3306)/dbname?charset=utf8' -o ./dao

Usage 2(Specified tables):
	go-dao-code-gen -h 127.0.0.1 -u root -p 123456 -D dbname -o ./dao -tables "tbl1,tbl2"

  -D string
    	Database to use.
  -P string
    	Port number to use for connection. (default "3306")
  -dsn string
    	Mysql dsn connection string.
  -h string
    	Connect to host. (default "127.0.0.1")
  -help
    	Show command usage.
  -o string
    	Output directory.
  -p string
    	Password to use when connecting to server.
  -params string
    	Connection parameters.
  -tables string
    	Generation range of tables, use "," separate multiple tables.
  -u string
    	User for login if not root user. (default "root")
  -v	Show command version.
```

## 生成的文件使用示例 ##

config.ini

Mysql config reference:[go-mysql](https://github.com/yangsai7/go-mysql)
``` toml
[mysql]
dsn = "test:test@(127.0.0.1:3306)/t_db"
retry = 2
db_conn_pool_max_idle = 100
db_conn_pool_max_open = 1000
db_conn_pool_max_lifetime = 10000000
```

main.go
``` go
package main

type Config struct {
	Mysql mysql.Config `toml:"mysql"`
}

var config Config
const ConfFile = "config.ini"

func main() {
	// init the config, only need init once in any project
	if _, err := toml.DecodeFile(ConfFile, &config); err != nil {
		fmt.Printf("fail to read config.||err=%v||config=%v", err, ConfFile)
		return
	}
	
    // init the mysql connection, only need init once in any project
	if err := dao.InitDao(&config.Mysql); err != nil {
		fmt.Printf("fail to get sql Factory instantce || err=%v", err)
        os.Exit(1)
	}
    defer dao.Close()
}
```

## shadow表支持 ##
go-dao-code-gen 会自动识别`fct_`开头和`_shadow`结尾的表名，并添加映射关系到[go-sqlbuilder](https://github.com/yangsai7/go-sqlbuilder)中。当请求的context带有压测标识，会自动把操作的表名替换成对应的shadow表名。
