package session

/*
 * 目前Go标准包没有为session提供支持，自行实现，最主要的三个问题：
 * 1. 生成全局唯一标识符（sessionid）
 * 2. 开辟数据存储空间
 * 3. 将session的全局唯一标示符发送给客户端
 *
 * 关于第三个问题，通常有两种方案：cookie和URL重写。
 * 1.Cookie：服务端通过设置Set-cookie头就可以将session的标识符传送到客户端，而客户端此后的每一次请求都会带上这个标识符
 * 2.URL重写：在返回给用户的页面里的所有的URL后面追加session标识符，这样用户在收到响应之后，无论点击响应页面里的哪个链接
 *           或提交表单，都会自动带上session标识符，如果客户端禁用了cookie的话，此种方案将会是首选。
 *
 * 本例采用方案1!
 */

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

func init() {
	fmt.Println("CCCCCCCCCCCCCCCCCC")
}

/*
 * session接口
 * 定义了Session的基本操作：set，get，delete，获取SESSIONID
 * 所以，只要实现了这4个方法的类型，就是一个Session类型
 */
type Session interface {
	Set(key, value interface{}) error //set session value
	Get(key interface{}) interface{}  //get session value
	Delete(key interface{}) error     //delete session value
	SessionID() string                //get current SESSIONID
}

/*
 * session是保存在服务器端的数据，可以以任何的方式存储，比如存储在内存、数据库或者文件
 * 因此抽象出一个Storage接口，每个实现了该接口的一个类型，就代表一种底层存储
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

//session管理器
type SessionManager struct {
	cookie_name   string     //cookie的名字（SessionId以cookie形式传到客户端）
	lock          sync.Mutex //protects session
	storager      Storage    //一种具体的存储实现
	max_life_time int64      //最大有效期，用于GC
}

/*
 * 参照database/sql/driver，先定义好接口，然后具体的存储session的结构只需要：1.实现相应的接口；2.注册，
 * 相应功能这样就可以使用了，g_storages全局变量负责存储全部的storage实现，Register函数负责填充g_storages
 */
var g_storages = make(map[string]Storage)

//创建管理器
func NewManager(storage_name, cookie_name string, max_life_time int64) (*SessionManager, error) {
	storager, ok := g_storages[storage_name]
	if !ok {
		return nil, fmt.Errorf("session: unknown storage %q (forgotten import?)", storage_name)
	}
	return &SessionManager{storager: storager, cookie_name: cookie_name, max_life_time: max_life_time}, nil
}

/*
 * Register函数，注册一个可用的storager。不能重复注册
 * 这个函数应该有各种storage的实现在其init中调用
 */
func Register(name string, storager Storage) {
	if storager == nil {
		panic("session: Register Storage is nil")
	}
	if _, dup := g_storages[name]; dup {
		panic("session: Register called twice for Storage " + name)
	}
	g_storages[name] = storager
}

//生成全局唯一的Session ID
func (manager *SessionManager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

//SessionStart函数：
//检查用户request的cookie中对应sid的值，如果没有，则创建sid；如果有，则读取session值，这个值又是什么呢？？？
func (manager *SessionManager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookie_name)
	if err != nil || cookie.Value == "" {
		sid := manager.sessionId() //生成全局唯一的sid
		session, _ = manager.storager.SessionInit(sid)
		cookie := http.Cookie{Name: manager.cookie_name, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(manager.max_life_time)}
		http.SetCookie(w, &cookie)
	} else {
		//cookie[cookie_name]对应的值，其实是sessionid!
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.storager.SessionRead(sid)
	}
	return
}

//Destroy session
//1. 服务端：先调用对应storager的sessiondestroy函数
//2. 客户端：然后让客户端清除cookie
func (manager *SessionManager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookie_name)
	if err != nil || cookie.Value == "" {
		return
	} else {
		manager.lock.Lock()
		defer manager.lock.Unlock()
		session_id := cookie.Value
		manager.storager.SessionDestroy(session_id)
		expiration := time.Now()
		// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
		cookie := http.Cookie{Name: manager.cookie_name, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

//利用了time包中的定时器功能，当超时maxLifeTime之后调用GC函数，这样就可以保证maxLifeTime时间内的session是可用的
//首先调用对应storager的GC，然后启动一个定时器，触发自己，也就是说manager的GC会定期触发~
func (manager *SessionManager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.storager.SessionGC(manager.max_life_time)
	time.AfterFunc(time.Duration(manager.max_life_time), func() { manager.GC() }) //自己调用自己
}
