package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	//此处使用相对路径，正式项目应该用绝对路径为佳
	"github.com/hq-cml/GoHttpWeb/practices/session/session"
	//"github.com/hq-cml/GoHttpWeb/practices/session/session/storages/memory"
	//"./session"
)

//全局的session管理器
var globalSessions *session.SessionManager

//包初始化函数
func init() {
	fmt.Println("Main init")
	globalSessions, _ = session.NewManager("memory", "GOSESSID", 3600)
	go globalSessions.GC()
}

//每当有客户访问login，就会有SessionStart，开始了奇幻之旅~
func login(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	r.ParseForm()
	if r.Method == "GET" {
		fmt.Println("First")
		t, _ := template.ParseFiles("login.gtpl")
		w.Header().Set("Content-Type", "text/html")
		t.Execute(w, sess.Get("username"))
	} else {
		fmt.Println("Not First")
		sess.Set("username", r.Form["username"])
		http.Redirect(w, r, "/", 302)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析url传递的参数，对于POST则解析响应包的主体（request body）
	//注意:如果没有调用ParseForm方法，下面无法获取表单的数据
	fmt.Println(r.Form) //这些信息是输出到服务器端的打印信息
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello world!") //这个写入到w的是输出到客户端的
}

func main() {
	http.HandleFunc("/login", login)         //设置访问的路由
	http.HandleFunc("/", hello)              //设置访问的路由
	err := http.ListenAndServe(":9527", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
