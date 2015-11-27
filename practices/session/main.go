package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	//此处使用相对路径，正式项目应该用绝对路径为佳
	//"github.com/hq-cml/GoHttpWeb/practices/session/session"
	"./session"
)

//全局的session管理器
var globalSessions *session.SessionManager

//包初始化函数
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
