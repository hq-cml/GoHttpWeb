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
	sid           string                      //session id唯一标示
	time_accessed time.Time                   //最后访问时间
	value         map[interface{}]interface{} //session里面存储的值
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
