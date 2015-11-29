package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	//此处使用相对路径，正式项目应该用绝对路径为佳
	"github.com/hq-cml/GoHttpWeb/practices/session/session"
	_ "github.com/hq-cml/GoHttpWeb/practices/session/session/storages"
	//"./session"
)

//全局的session管理器
var g_sessions *session.SessionManager

//包初始化函数
func init() {
	fmt.Println("Main init")
	g_sessions, _ = session.NewManager("memory", "GOSESSID", 3600)
	go g_sessions.GC()
}

//每当有客户访问login，就会有SessionStart，开始了奇幻之旅~
func login(w http.ResponseWriter, r *http.Request) {
	sess := g_sessions.SessionStart(w, r)
	r.ParseForm()
	//如果是从表单提交过来的访问，method应该是post，如果是直接浏览器访问，则是get
	if r.Method == "GET" {
		fmt.Println("First com")
		t, _ := template.ParseFiles("login.gtpl")
		w.Header().Set("Content-Type", "text/html")
		t.Execute(w, sess.Get("username"))
		fmt.Printf("Login Session: %+v\n", sess.Get("username"))
	} else {
		fmt.Println("Not First")
		sess.Set("username", r.Form["username"])
		http.Redirect(w, r, "/", 302)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	sess := g_sessions.SessionStart(w, r)
	if sess.Get("username") == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	fmt.Printf("Hello Session: %+v\n", sess.Get("username"))
	fmt.Fprintf(w, "Hello %+v!", sess.Get("username")) //这个写入到w的是输出到客户端的
}

func main() {
	http.HandleFunc("/login", login)         //设置访问的路由
	http.HandleFunc("/", hello)              //设置访问的路由
	err := http.ListenAndServe(":9527", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
