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
	lock     sync.Mutex               //锁
	sessions map[string]*list.Element //用于存储的内存，key是sid，value是list的Element（其实本质上，是一个）
	list     *list.List               //链表，用于gc
}

var g_redis_storage = &RedisStorage{}
var g_redis_client goredis.Client

func init() {
	fmt.Println("Redis storage init")
	// 设置端口为redis默认端口
	g_redis_client.Addr = "127.0.0.1:6379"
	g_redis_storage.list = list.New()
	g_redis_storage.sessions = make(map[string]*list.Element, 0)
	session.Register("redis", g_redis_storage)
}

/*
 * RedisSession实现Session接口的：Set/Get/Delete/SessionID方法
 */
func (self *RedisSession) Set(key, value interface{}) error {
	var k string
	var v []byte
	var ok bool
	if k, ok = key.(string); !ok {
		return nil
	}
	if v, ok = value.([]byte); !ok {
		return nil
	}
	g_redis_client.Hset(self.sid, k, v)
	//更新对应条目的访问时间
	g_redis_storage.SessionUpdate(self.sid)
	return nil
}

func (self *RedisSession) Get(key interface{}) interface{} {
	var k string
	var ok bool
	if k, ok = key.(string); !ok {
		return nil
	}
	//更新对应条目的访问时间
	g_redis_storage.SessionUpdate(self.sid)
	if v, err := g_redis_client.Hget(self.sid, k); err != nil {
		return v
	} else {
		return nil
	}
	return nil
}

func (self *RedisSession) Delete(key interface{}) error {
	var k string
	var ok bool
	if k, ok = key.(string); !ok {
		return nil
	}
	//更新对应条目的访问时间
	g_redis_storage.SessionUpdate(self.sid)
	g_redis_client.Hdel(self.sid, k)
	return nil
}

func (self *RedisSession) SessionID() string {
	return self.sid
}

/*
 * RedisStorage实现Storage接口的：SessionInit/SessionFetch/SessionDestroy/SessionGC方法
 */
//当新来一个用户的时候，新增一个session条目（element），并将之返回
func (self *RedisStorage) SessionInit(sid string) (session.Session, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	newsess := &MemSession{sid: sid, time_accessed: time.Now()}
	//将新生成的条目压入队列，开始GC轮回
	element := self.list.PushBack(newsess)
	//将新生成的条目以element的形式，放入session中去，用于后续读写
	self.sessions[sid] = element
	return newsess, nil
}

//根据sid，从storage中取出整个对应的条目（Element），以MemSession形式返回
func (self *RedisStorage) SessionFetch(sid string) (session.Session, error) {
	if element, ok := self.sessions[sid]; ok {
		return element.Value.(*RedisSession), nil
	} else {
		sess, err := self.SessionInit(sid)
		return sess, err
	}
	return nil, nil
}

//根据sid，销毁storage中对应的条目，两处，内存中和gc队列中均需要清除
func (self *RedisStorage) SessionDestroy(sid string) error {
	if element, ok := self.sessions[sid]; ok {
		delete(self.sessions, sid)
		self.list.Remove(element)
		g_redis_client.Del(sid)
		return nil
	}
	return nil
}

//GC，从最久未被访问的条目，一直向前遍历。
//如果条目的访问时间+max_life_time比当前时间还小，则表示过期，则在队列以及内存中均予以删除
func (self *RedisStorage) SessionGC(max_life_time int64) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for {
		element := self.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*MemSession).time_accessed.Unix() + max_life_time) < time.Now().Unix() {
			self.list.Remove(element)
			delete(self.sessions, element.Value.(*MemSession).sid)
			g_redis_client.Del(element.Value.(*MemSession).sid)
		} else {
			break
		}
	}
}

//跟新session存储中sid对应的条目（element）的更新时间，并且将对应条目前移
func (self *RedisStorage) SessionUpdate(sid string) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if element, ok := self.sessions[sid]; ok {
		element.Value.(*RedisSession).time_accessed = time.Now()
		self.list.MoveToFront(element)
		return nil
	}
	return nil
}
