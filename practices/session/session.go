package session

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

//session管理器
type SessionManager struct {
	cookieName  string     //private cookiename
	lock        sync.Mutex // protects session
	provider    Provider
	maxlifetime int64
}

/*
 * session是保存在服务器端的数据，可以以任何的方式存储，比如存储在内存、数据库或者文件
 * 因此抽象出一个Provider接口，用以表征session管理器底层存储结构
 *
 * SessionInit函数实现Session的初始化，操作成功则返回此新的Session变量
 * SessionRead函数返回sid所代表的Session变量，如果不存在，那么将以sid为参数调用SessionInit函数创建并返回一个新的Session变量
 * SessionDestroy函数用来销毁sid对应的Session变量
 * SessionGC根据maxLifeTime来删除过期的数据
 */
type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

//创建管理器
func NewManager(provideName, cookieName string, maxlifetime int64) (*SessionManager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	return &SessionManager{provider: provider, cookieName: cookieName, maxlifetime: maxlifetime}, nil
}
