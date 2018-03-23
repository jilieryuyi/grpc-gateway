#!/usr/bin/env bash
echo "正在编译..."
current_path=$(cd `dirname $0`; pwd)

##/Users/yuyi/go  is GOROOT 修改这个路径为自己的GOPATH就可以了
export GOPATH="/Users/yuyi/go:${current_path}"

##进入当前目录
cd ${current_path}
##build构建项目
go build -p 4 -race server ##-a强制重新编译所有的包 -v显示被编译的包 -x显示所用到的其他命令
##install安装
go install server