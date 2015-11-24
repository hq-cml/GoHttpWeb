package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	//设置cookie
	expiration := time.Now()
	expiration = expiration.AddDate(1, 0, 0)
	cookie := http.Cookie{Name: "username", Value: "hq", Expires: expiration}
	http.SetCookie(w, &cookie)

	//读取cookie
	fmt.Fprintf(w, "Hello world~") //这个写入到w的是输出到客户端的
}

func main() {
	http.HandleFunc("/", sayhelloName)       //设置访问的路由
	err := http.ListenAndServe(":9527", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
