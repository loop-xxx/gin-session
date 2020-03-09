package main

import (
	"github.com/gin-gonic/gin"
	gin_session "github.com/loop-xxx/gin-session"
	"github.com/loop-xxx/gin-session/dao"
	"net/http"
)

type User struct{
	FirstName string
	SecondName string
}

func main() {
	if keeper, err := dao.DefaultRedis("192.168.20.130:6379", "toor", 0); err == nil {
		engine := gin.Default()
		engine.Use(gin_session.DefaultGinSessionManager(keeper, "localhost"))

		engine.GET("/login", func(ctx *gin.Context){
			if session, exist := gin_session.GetSession(ctx); exist {
				session.Set("name", "loop")
				_= session.SetStruct("user", User{"li", "loop"})
			}
			ctx.String(http.StatusOK, "ok")
		})


		engine.GET("/show", func(ctx *gin.Context){
			if session, ok := gin_session.GetSession(ctx); ok {
				if name , ok := session.Get("name"); ok{
					var u User
					if err := session.GetStruct("user",&u); err == nil{
						ctx.JSON(http.StatusOK, gin.H{"name": name,  "user":u})
					}
				}
			}
		})
		_ = engine.Run(":2333")
	}
}
