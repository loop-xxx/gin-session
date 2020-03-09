package gin_session

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/loop-xxx/gin-session/cache_pool"
	"github.com/loop-xxx/gin-session/dao"
	uuid "github.com/satori/go.uuid"
	"time"
)

func GinSessionManager(keeper dao.Keeper, cookieMaxAge, expiration time.Duration, domain string, poolMaxSize, sessionInitSize int)func (gin.Context)(err error){
	pool := cache_pool.NewCachePool(poolMaxSize)
	return func(ctx *gin.Context){
		var data map[string]string
		var ball *cache_pool.CacheBall

		if token, err := ctx.Cookie("gin-session-id"); err == nil{

			if cacheBall, exist := pool.SearchCacheBall(token); exist {
				ball = cacheBall
			}else{
				//该token对应cache ball已经被lru算法移除, 为该token创建新的cache bool并托管到cache pool
				ball = cache_pool.MakeCacheBall(fmt.Sprintf("gin-session:%s", token), keeper, expiration)
				pool.AppendCacheBall(token, ball)
			}

			if ball.Sync(){
				data = ball.Get()
			}else{
				//同步失败的原因:
				//keeper (redis)中的key-value不存在或已过期
				//keeper 底层错误(redis-server连接错误)
				//keeper 的实现代码有bug
				data = make(map[string]string, sessionInitSize)
			}
		}else {
			// 如果本次请求是用户端第一次请求该网站那么拿不到token, 需要为用户创建新的token
			token := uuid.NewV4().String()
			ctx.SetCookie("gin-session-id", token,  cookieMaxAge, "/",domain,
				false, //是否只支持https
				true)//是否不支持js访问

			//为新token创建新的cache bool并托管到cache pool
			ball = cache_pool.MakeCacheBall(fmt.Sprintf("gin-session:%s", token), keeper, expiration)
			pool.AppendCacheBall(token, ball)

			data = make(map[string]string, sessionInitSize)
		}

		ginSession := New(data)
		ctx.Set("gin-session", ginSession)
		ctx.Next()
		if ginSession.Check(){
			ball.SetAndCommit(ginSession.Dump())
		}
	}
}

func GetSession(ctx *gin.Context)(s Session, exist bool){
	if ginSession, exists := ctx.Get("gin-session"); exists{
		s = ginSession.(Session)
		exist = true
	}
	return
}

