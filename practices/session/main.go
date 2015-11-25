package main

import (
	"fmt"
	"log"
	"net/http"
	"session"
	"time"
)

//全局的session管理器
var globalSessions *session.SessionManager

func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
}

func main() {
	http.HandleFunc("/", sayhelloName)       //设置访问的路由
	err := http.ListenAndServe(":9527", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
