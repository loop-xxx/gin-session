package session

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/loop-xxx/gin-session/cache"
	"github.com/loop-xxx/gin-session/dao"
	uuid "github.com/satori/go.uuid"
	"time"
)


func DefaultGinSessionManager(keeper dao.Keeper, domain string)func (*gin.Context){
	return GinSessionManager(
		keeper, //底层为一个redis客户端
		domain, //web服务器的域名
		time.Minute*30, //session的超时时间 30 分钟
		0x100, //cache-pool的最大容积
		0x4, //一个空的session被初始化时的最初打下
		)
}

func GinSessionManager(keeper dao.Keeper, domain string,
	expiration time.Duration, poolMaxSize, sessionMapInitSize int)func (*gin.Context){
	pool := cache.NewCachePool(poolMaxSize)
	return func(ctx *gin.Context){
		var data map[string]string
		var ball *cache.CacheBall

		//1 获取请求携带的session
		if token, err := ctx.Cookie("gin-session-id"); err == nil{
			//2 到pool中查找有没有对应的ball
			if cacheBall, exist := pool.SearchCacheBall(token); exist {
				ball = cacheBall
			}else{
				//3 如果未找到则创建新的ball, 并将其交给pool拓展
				//该token对应cache ball已经被lru算法移除, 为该token创建新的cache bool并托管到cache pool
				ball = cache.MakeCacheBall(fmt.Sprintf("gin-session:%s", token), keeper, expiration)
				pool.AppendCacheBall(token, ball)
			}

			//4 获取数据前, 先于redis同步以下数据
			if ball.Sync(){
				//5 同步成功后获取一个cache-ball保存的数据副本
				data = ball.Get()
			}else{
				//同步失败的原因:
				//keeper (redis)中的key-value不存在或已过期
				//keeper 底层错误(redis-server连接错误)
				//keeper 的实现代码有bug

				//未读取到任何数据则, 创建一个空的session-map
				data = make(map[string]string, sessionMapInitSize)
			}
		}else {
			// 如果本次请求是用户端第一次请求该网站那么拿不到token, 需要为用户创建新的token
			token := uuid.NewV4().String()
			//为新token创建新的cache bool并托管到cache pool
			ball = cache.MakeCacheBall(fmt.Sprintf("gin-session:%s", token), keeper, expiration)
			pool.AppendCacheBall(token, ball)

			//设置客户端的session-id
			ctx.SetCookie("gin-session-id", token,  0, "/",domain,
				false, //是否只支持https
				true)//是否不支持js访问

			//创建空的session-map
			data = make(map[string]string, sessionMapInitSize)
		}

		ginSession := New(data)
		ctx.Set("gin-session", ginSession)

		ctx.Next()
		//请求结束后判断副本数据是否被修改过, 若修改过提交到cache-ball, cache-ball也会主动同步到redis
		if ginSession.Check(){
			ball.Commit(ginSession.Dump())
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

