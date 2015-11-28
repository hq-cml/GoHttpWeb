package session

import (
	"container/list"
	"sync"
	"time"
)

/*
 * 内存Session实现，这个结构实现Session接口
 */
type MemSession struct {
	sid          string                      //session id唯一标示
	timeAccessed time.Time                   //最后访问时间
	value        map[interface{}]interface{} //session里面存储的值
}

/*
 * 内存存储实现，这个结构实现Storage接口
 */
type MemStorage struct {
	lock     sync.Mutex               //锁
	sessions map[string]*list.Element //用于存储的内存
	list     *list.List               //链表，用用于gc
}

var g_memstorage = &MemStorage{list: list.New()}

func init() {
	g_memstorage.sessions = make(map[string]*list.Element, 0)
	Register("memory", g_memstorage)
}

func (st *MemSession) Set(key, value interface{}) error {
	st.value[key] = value
	g_memstorage.SessionUpdate(st.sid)
	return nil
}

func (st *MemSession) Get(key interface{}) interface{} {
	g_memstorage.SessionUpdate(st.sid)
	if v, ok := st.value[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (st *MemSession) Delete(key interface{}) error {
	delete(st.value, key)
	g_memstorage.SessionUpdate(st.sid)
	return nil
}

func (st *MemSession) SessionID() string {
	return st.sid
}

func (self *MemStorage) SessionInit(sid string) (Session, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := self.list.PushBack(newsess)
	self.sessions[sid] = element
	return newsess, nil
}

func (self *MemStorage) SessionRead(sid string) (Session, error) {
	if element, ok := self.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		sess, err := self.SessionInit(sid)
		return sess, err
	}
	return nil, nil
}

func (self *MemStorage) SessionDestroy(sid string) error {
	if element, ok := self.sessions[sid]; ok {
		delete(self.sessions, sid)
		self.list.Remove(element)
		return nil
	}
	return nil
}

func (self *MemStorage) SessionGC(maxlifetime int64) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for {
		element := self.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			self.list.Remove(element)
			delete(self.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

func (self *MemStorage) SessionUpdate(sid string) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if element, ok := self.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		self.list.MoveToFront(element)
		return nil
	}
	return nil
}
