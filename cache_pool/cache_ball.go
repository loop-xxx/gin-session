package cache_pool

import (
	"github.com/loop-xxx/gin-session/dao"
	"log"
	"sync"
	"time"
)

type CacheBall struct {
	//读写锁
	rwm sync.RWMutex

	//redis客户端
	keeper dao.Keeper
	//每次访问redis中的数据, 都会将数据的存活时间重置为该事件
	expiration time.Duration
	//数据在redis中的key
	key string

	//当前缓存数据版本的唯一标示
	magic string
	//缓存数据
	dataMap map[string]string
}

func (cnp *CacheBall) _syncIn() (success bool) {
	cnp.rwm.Lock()
	defer cnp.rwm.Unlock()

	//调用Peek函数再次校验数据是否需要重新拉取, 以防并发时重复拉取数据
	if different, err := cnp.keeper.Peek(cnp.key, cnp.magic); err == nil {
		if different {
			if m, dm, err := cnp.keeper.Pull(cnp.key); err == nil {
				//拉取数据,更新magic和数据
				cnp.magic = m
				cnp.dataMap = dm
				success = true
			} else {
				log.Printf("[cache-pool ERROR] %v\n", err)
			}
		} else {
			success = true
		}
	} else {
		log.Printf("[cache-pool ERROR] %v\n", err)
	}
	return
}
func (cnp *CacheBall) Sync() (success bool) {
	//通过magic 检测Redis中的key有没有超时, Session缓存有没有过期
	cnp.rwm.RLock()
	status, err := cnp.keeper.Check(cnp.key, cnp.magic, cnp.expiration)
	cnp.rwm.RUnlock()

	if err == nil {
		switch status {
		case dao.GinSessionOk:
			success = true
		case dao.GinSessionOld:
			success = cnp._syncIn()
		}
	} else {
		log.Printf("[cache-pool ERROR] %v\n", err)
	}
	return
}

func (cnp *CacheBall) Get() (dataMap map[string]string) {
	cnp.rwm.RLock()
	defer cnp.rwm.RUnlock()

	dataMap = make(map[string]string, len(cnp.dataMap))
	for key, value := range cnp.dataMap {
		dataMap[key] = value
	}
	return dataMap
}

func (cnp *CacheBall) _syncOut() (success bool) {
	success = true
	//提交数据到redis
	if err := cnp.keeper.Push(cnp.key, cnp.magic, cnp.dataMap, cnp.expiration); err != nil {
		success = false
		// module层出错
		log.Printf("[cache-pool ERROR] %v\n", err)
	}
	return
}

func (cnp *CacheBall) Commit(dataMap map[string]string) (success bool) {
	//生成数据当前版本的magic

	magicInt64 := time.Now().UnixNano()

	magicBytes := [4]byte{}
	magicBytes[3] = uint8(magicInt64 >> 18)
	magicBytes[2] = uint8(magicInt64 >> 26)
	magicBytes[1] = uint8(magicInt64 >> 34)
	magicBytes[0] = uint8(magicInt64 >> 42)

	magic := string(magicBytes[:])
	cnp.rwm.Lock()
	defer cnp.rwm.Unlock()

	if cnp.dataMap == nil{
		cnp.dataMap = make(map[string]string , len(dataMap))
	}
	cnp.magic = magic
	for key, value := range dataMap {
		cnp.dataMap[key] = value
	}

	success = cnp._syncOut()
	return
}

func MakeCacheBall(key string, keeper dao.Keeper, expiration time.Duration) (cacheBall *CacheBall) {
	cacheBall = &CacheBall{
		rwm:        sync.RWMutex{},
		keeper:     keeper,
		expiration: expiration,
		key:        key,
		magic:      "none", //default magic
		dataMap:    nil,
	}
	return
}
