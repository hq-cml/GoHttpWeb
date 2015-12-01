package storages

import (
	"container/list"
	"fmt"
	"github.com/astaxie/goredis"
	"github.com/hq-cml/GoHttpWeb/practices/session/session"
	"sync"
	"time"
)

/*
 * RedisSession实现，实现Session接口
 * 这是一个用户对应的session结构，而不是整体的session结构
 */
type RedisSession struct {
	sid           string    //session id唯一标示
	time_accessed time.Time //最后访问时间
	//value         map[interface{}]interface{} //session里面存储的值
}

/*
 * Redis存储实现，这个结构实现Storage接口
 * 这是一个整体session的对应的结构
 */
type RedisStorage struct {
	redis_client goredis.Client
	lock         sync.Mutex //锁
	//sessions map[string]*list.Element //用于存储的内存，key是sid，value是list的Element（其实本质上，是一个）
	list *list.List //链表，用于gc
}

var g_redis_storage = &RedisStorage{}

func init() {
	fmt.Println("Redis storage init")
	// 设置端口为redis默认端口
	g_redis_storage.redis_client.Addr = "127.0.0.1:6379"
	g_redis_storage.list = list.New()
	//g_redis_storage.sessions = make(map[string]*list.Element, 0)
	session.Register("redis", g_redis_storage)
}

/*
 * RedisSession实现Session接口的：Set/Get/Delete/SessionID方法
 */
func (self *RedisSession) Set(key, value interface{}) error {
	self.redis_client.Hset(self.sid, key, []byte(value))
	//更新对应条目的访问时间
	g_redis_storage.SessionUpdate(self.sid)
	return nil
}

func (self *RedisSession) Get(key interface{}) interface{} {
	//更新对应条目的访问时间
	g_redis_storage.SessionUpdate(self.sid)
	if v, ok := self.redis_client.Hget(self.sid, key); ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (self *RedisSession) Delete(key interface{}) error {
	//更新对应条目的访问时间
	g_redis_storage.SessionUpdate(self.sid)
	self.redis_client.Hdel(self.sid)
	return nil
}

func (self *MemSession) SessionID() string {
	return self.sid
}
