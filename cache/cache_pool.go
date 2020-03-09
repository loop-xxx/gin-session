package cache

import "sync"

type cacheNode struct{
	token string
	prev *cacheNode
	next *cacheNode

	core *CacheBall
}
func _link(prev , next, cn *cacheNode){
	prev.next = cn
	cn.prev = prev

	next.prev = cn
	cn.next = next
}
func _unlink(prev, next *cacheNode){
	prev.next = next
	next.prev = prev
}

func first(head *cacheNode, cn *cacheNode){
	_unlink(cn.prev, cn.next)
	_link(head, head.next, cn)
}
func push(head *cacheNode, cn *cacheNode){
	_link(head, head.next, cn)
}
func del( cn *cacheNode){
	_unlink(cn.prev, cn.next)
}


type CachePool struct{
	//同步锁
	rwm sync.RWMutex

	//用于lru算法
	lru  *cacheNode
	//用于根据token找到对应cacheNode
	hash map[string]*cacheNode

	//cachePool的最大容量
	maxSize int
}

func NewCachePool(maxSize int)(cachePool *CachePool){
	//初始化头节点
	head := &cacheNode{}
	head.prev = head
	head.next = head

	//初始化cachePool
	cachePool = &CachePool{
		rwm:     sync.RWMutex{},
		lru:     head,
		hash:    make(map[string]*cacheNode, maxSize/2),
		maxSize: maxSize,
	}

	return
}


func (cpp *CachePool)SearchCacheBall(token string)(cacheBall *CacheBall, exist bool){
	cpp.rwm.RLock()
	defer cpp.rwm.RUnlock()

	if cacheNode, ok := cpp.hash[token]; ok{
		//进入修改模式
		cpp.rwm.RUnlock()
		cpp.rwm.Lock()

		first(cpp.lru, cacheNode)

		//推出修改模式
		cpp.rwm.Unlock()
		cpp.rwm.RLock()

		exist = true
		cacheBall = cacheNode.core
	}
	return
}


func (cpp *CachePool)AppendCacheBall(token string, cacheBall *CacheBall){
	cpp.rwm.Lock()
	defer cpp.rwm.Unlock()

	if len(cpp.hash) >= cpp.maxSize{
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