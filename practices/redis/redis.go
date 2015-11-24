package main

import (
	"fmt"
	"github.com/astaxie/goredis"
)

func main() {
	var client goredis.Client
	// 设置端口为redis默认端口
	client.Addr = "127.0.0.1:6379"

	//字符串操作
	client.Set("a", []byte("hello"))
	val, _ := client.Get("a")
	fmt.Println(string(val))
	client.Del("a")
}
