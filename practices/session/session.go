package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
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
	storager    Storage
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
type Storage interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

/*
 * session接口定义
 * Session的处理基本就 设置值、读取值、删除值以及获取当前sessionID这四个操作
 * 所以Session接口实现这四个操作
 */
type Session interface {
	Set(key, value interface{}) error //set session value
	Get(key interface{}) interface{}  //get session value
	Delete(key interface{}) error     //delete session value
	SessionID() string                //back current sessionID
}

//创建管理器
func NewManager(storageName, cookieName string, maxlifetime int64) (*SessionManager, error) {
	storager, ok := storages[storageName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", storageName)
	}
	return &SessionManager{storager: storager, cookieName: cookieName, maxlifetime: maxlifetime}, nil
}

/*
 * 参照database/sql/driver，先定义好接口，然后具体的存储session的结构只需要：1.实现相应的接口；2.注册，
 * 相应功能这样就可以使用了，storages全局变量负责存储全部的storager实现，Register函数负责填充storages
 */
var storages = make(map[string]Storage)

/*
 * Register函数，注册一个可用的storager。不能重复注册
 */
func Register(name string, storager Storage) {
	if storager == nil {
		panic("session: Register Storage is nil")
	}
	if _, dup := storages[name]; dup {
		panic("session: Register called twice for Storage " + name)
	}
	storages[name] = storager
}

//生成全局唯一的Session ID
func (manager *SessionManager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

//SessionStart函数，检测是否已经有某个Session与当前来访用户发生了关联，如果没有则创建之。
func (manager *SessionManager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sid := manager.sessionId()
		session, _ = manager.provider.SessionInit(sid)
		cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(manager.maxlifetime)}
		http.SetCookie(w, &cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRead(sid)
	}
	return
}
