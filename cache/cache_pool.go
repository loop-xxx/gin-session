package cache

import "sync"

type cacheNode struct {
	token string
	prev  *cacheNode
	next  *cacheNode

	core *Ball
}

func _link(prev, next, cn *cacheNode) {
	prev.next = cn
	cn.prev = prev

	next.prev = cn
	cn.next = next
}
func _unlink(prev, next *cacheNode) {
	prev.next = next
	next.prev = prev
}

func first(head *cacheNode, cn *cacheNode) {
	_unlink(cn.prev, cn.next)
	_link(head, head.next, cn)
}
func push(head *cacheNode, cn *cacheNode) {
	_link(head, head.next, cn)
}
func del(cn *cacheNode) {
	_unlink(cn.prev, cn.next)
}

type Pool struct {
	//同步锁
	mx sync.Mutex
	//用于lru算法, 双向循环链表
	lru *cacheNode
	//用于根据token找到对应cacheNode, hash表
	hash map[string]*cacheNode
	//cachePool的最大容量
	maxSize int
}

func (cpp *Pool) SearchCacheBall(token string) (cacheBall *Ball, exist bool) {
	cpp.mx.Lock()
	defer cpp.mx.Unlock()

	if cacheNode, ok := cpp.hash[token]; ok {
		first(cpp.lru, cacheNode)
		exist = true
		cacheBall = cacheNode.core
	}
	return
}

func (cpp *Pool) AppendCacheBall(token string, cacheBall *Ball) {
	cpp.mx.Lock()
	defer cpp.mx.Unlock()

	if len(cpp.hash) >= cpp.maxSize {
		//如果CachePool已经满了,则删除池中最不常用的cacheNode
		delete(cpp.hash, cpp.lru.prev.token)
		del(cpp.lru.prev)
	}
	cacheNode := &cacheNode{
		token: token,
		core:  cacheBall,
	}
	cpp.hash[token] = cacheNode
	push(cpp.lru, cacheNode)
}

func NewCachePool(maxSize int) (cachePool *Pool) {
	//初始化头节点
	head := &cacheNode{}
	head.prev = head
	head.next = head

	//初始化cachePool
	cachePool = &Pool{
		mx:      sync.Mutex{},
		lru:     head,
		hash:    make(map[string]*cacheNode, maxSize/2),
		maxSize: maxSize,
	}

	return
}
