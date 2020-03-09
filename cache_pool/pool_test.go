package cache_pool

import (
	"fmt"
	"github.com/loop-xxx/gin-session/dao"
	"testing"
	"time"
)

func TestMakeCacheBall(t *testing.T) {
	if keeper, err := dao.DefaultRedis("192.168.20.130:6379", "toor", 0) ;err == nil{
		cacheBall := MakeCacheBall("uuid", keeper, time.Second*100)

		//提交失败的原因：
		//keeper底层错误(redis-server连接错误)
		//keeper的实现代码有bug
		if cacheBall.Commit(map[string]string{"name": "loop"}) {
			if cacheBall.Sync() {
				fmt.Println(cacheBall.Get())
			}else{
				//同步失败的原因:
				//keeper(redis)中的key-value不存在或已过期
				//keeper底层错误(redis-server连接错误)
				//keeper的实现代码有bug
				fmt.Println("Sync failed")
			}
		}else{
			fmt.Println("Commit failed")
		}
	}else{
		t.Error(err)
	}
}

func TestNewCachePool(t *testing.T) {
	pool := NewCachePool(0x4)
	for i := 0 ; i < 0x6 ; i ++{
		pool.AppendCacheBall(fmt.Sprintf("ll%d", i), nil )
	}
	for i:= 0x5; i >= 0; i--{
		if _, exist := pool.SearchCacheBall(fmt.Sprintf("ll%d", i)); !exist{
			fmt.Printf("lru delete %d\n", i)
		}
	}
}