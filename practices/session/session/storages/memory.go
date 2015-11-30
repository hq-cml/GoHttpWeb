package storages

import (
	"container/list"
	"fmt"
	"github.com/hq-cml/GoHttpWeb/practices/session/session"
	"sync"
	"time"
)

/*
 * 内存Session实现，这个结构实现Session接口
 * 这个更准确的说是一个用户对应的session结构，而不是整体的session结构
 */
type MemSession struct {
	sid           string                      //session id唯一标示
	time_accessed time.Time                   //最后访问时间
	value         map[interface{}]interface{} //session里面存储的值
}

/*
 * 内存存储实现，这个结构实现Storage接口
 * 这是一个整体session的对应的结构
 */
type MemStorage struct {
	lock     sync.Mutex               //锁
	sessions map[string]*list.Element //用于存储的内存，key是sid，value是list的Element（其实本质上，是一个）
	list     *list.List               //链表，用于gc
}

var g_memstorage = &MemStorage{}

func init() {
	fmt.Println("Mem storage init")
	g_memstorage.list = list.New()
	g_memstorage.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", g_memstorage)
}

/*
 * MemSession实现Session接口的：Set/Get/Delete/SessionID方法
 */
func (self *MemSession) Set(key, value interface{}) error {
	self.value[key] = value
	//更新对应条目的访问时间
	g_memstorage.SessionUpdate(self.sid)
	return nil
}

func (self *MemSession) Get(key interface{}) interface{} {
	//更新对应条目的访问时间
	g_memstorage.SessionUpdate(self.sid)
	if v, ok := self.value[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (self *MemSession) Delete(key interface{}) error {
	delete(self.value, key)
	g_memstorage.SessionUpdate(self.sid)
	return nil
}

func (self *MemSession) SessionID() string {
	return self.sid
}

/*
 * MemStorage实现Storage接口的：SessionInit/SessionFetch/SessionDestroy/SessionGC方法
 */
//当新来一个用户的时候，新增一个session条目（element）
func (self *MemStorage) SessionInit(sid string) (session.Session, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &MemSession{sid: sid, time_accessed: time.Now(), value: v}
	//将新生成的条目压入队列，开始GC轮回
	element := self.list.PushBack(newsess)
	//将新生成的条目以element的形式，放入session中去，用于后续读写
	self.sessions[sid] = element
	return newsess, nil
}

//根据sid，从storage中取出整个对应的条目（Element），以MemSession形式返回
func (self *MemStorage) SessionFetch(sid string) (session.Session, error) {
	if element, ok := self.sessions[sid]; ok {
		return element.Value.(*MemSession), nil
	} else {
		sess, err := self.SessionInit(sid)
		return sess, err
	}
	return nil, nil
}

//根据sid，销毁storage中对应的条目，两处，内存中和gc队列中均需要清除
func (self *MemStorage) SessionDestroy(sid string) error {
	if element, ok := self.sessions[sid]; ok {
		delete(self.sessions, sid)
		self.list.Remove(element)
		return nil
	}
	return nil
}

//GC，从最久未被访问的条目，一直向前遍历。
//如果条目的访问时间+max_life_time比当前时间还小，则表示过期，则在队列以及内存中均予以删除
func (self *MemStorage) SessionGC(max_life_time int64) {
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
		} else {
			break
		}
	}
}

//跟新session存储中sid对应的条目（element）的更新时间，并且将对应条目前移
func (self *MemStorage) SessionUpdate(sid string) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if element, ok := self.sessions[sid]; ok {
		element.Value.(*MemSession).time_accessed = time.Now()
		self.list.MoveToFront(element)
		return nil
	}
	return nil
}
