package dao

import "time"

type Status int

const (
	GinSessionOk = Status(iota)
	GinSessionOld
	GinSessionTimeout
)

func (s Status)String()(str string){
	switch s {
	case GinSessionOk:
		str = "gin-session_ok"
	case GinSessionOld:
		str = "gin-session_old"
	case GinSessionTimeout:
		str = "gin-session_timeout"
	default:
		str = "gin-session_unknown"
	}
	return
}

type Keeper interface {
	//首先, 执行expire 刷新key的存活时间如果失败说明该key已经超时失效
	//expire 执行成功后, 再执行Get获取magic进行校验
	Check(key, magic string, expiration time.Duration) (Status, error)
	//Peek不会刷新key的存活时间, 直接执行Get获取magic进行校验, 需要同步返回true, 不需要返回false
	Peek(key, magic string) (bool, error)
	//执行Pull函数时, 肯定校验过, Key的存活时间已经被刷新过了, 不用担心key已经超时的问题, 直接HMGet拉取数据
	Pull(key string)(magic string ,dataMap map[string]string,  err error)
	//首先,执行HMSet写入数据, redis set后默认key的存活时间为永久
	//然后,expire 设置key的存活时间
	Push(key string, magic string, dataMap map[string]string, expiration time.Duration) error
	//关闭redis-client
	Done() error
}
