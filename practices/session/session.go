package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

/*
 * 目前Go标准包没有为session提供任何支持，所以，自己实现，最主要的三个问题：
 * 1. 生成全局唯一标识符（sessionid）；
 * 2. 开辟数据存储空间。
 * 3. 将session的全局唯一标示符发送给客户端。
 *
 * 关于第三个问题，通常有两种方案：cookie和URL重写。
 * Cookie：服务端通过设置Set-cookie头就可以将session的标识符传送到客户端，而客户端此后的每一次请求都会带上这个标识符
 * URL重写：在返回给用户的页面里的所有的URL后面追加session标识符，这样用户在收到响应之后，无论点击响应页面里的哪个链接或提交表单，都会自动带上session标识符，如果客户端禁用了cookie的话，此种方案将会是首选。
 */

//全局的session管理器
type SessionManager struct {
	cookieName  string     //private cookiename
	lock        sync.Mutex // protects session
	provider    Provider
	maxlifetime int64
}

//创建全局管理器
func NewManager(provideName, cookieName string, maxlifetime int64) (*SessionManager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	return &SessionManager{provider: provider, cookieName: cookieName, maxlifetime: maxlifetime}, nil
}

var globalSessions *SessionManager

func init() {
	globalSessions, _ = NewManager("memory", "gosessionid", 3600)
}

func main() {
	http.HandleFunc("/", sayhelloName)       //设置访问的路由
	err := http.ListenAndServe(":9527", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
