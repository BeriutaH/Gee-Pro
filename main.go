package main

import (
	"gee"
	"net/http"
)

func main() {
	r := gee.New() // 全局共享一个Engine
	//r.GET("/", func(c *gee.Context) {
	//	c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	//})
	//
	//r.GET("/hello", func(c *gee.Context) {
	//	c.String(http.StatusOK, "hello %s, you are at %s\n", c.Query("name"), c.Path)
	//})
	//
	//r.GET("/hello/:name", func(c *gee.Context) {
	//	log.Printf("name %s", c.Param("name"))
	//	c.String(http.StatusOK, "hello %s, you are at %s\n", c.Param("name"), c.Path)
	//})
	//
	//r.POST("/assets/*filepath", func(c *gee.Context) {
	//
	//	c.JSON(http.StatusOK, gee.H{
	//		"username": c.PostForm("username"),
	//		"password": c.PostForm("password"),
	//		"filepath": c.Param("filepath"),
	//	})
	//})

	// 路由组测试
	r.GET("/index", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/", func(ctx *gee.Context) {
			ctx.HTML(http.StatusOK, "<h1>Hello Home</h1>")
		})
		v1.GET("/hello", func(ctx *gee.Context) {
			ctx.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

	}

	v2 := r.Group("/admin")
	{
		v2.POST("/login", func(ctx *gee.Context) {
			_ = ctx.JSON(http.StatusOK, gee.H{
				"username": ctx.PostForm("username"),
				"password": ctx.PostForm("password"),
			})
		})
		v2.POST("/info/:name", func(ctx *gee.Context) {
			_ = ctx.JSON(http.StatusOK, gee.H{
				"username": ctx.Param("name"),
			})
		})
	}
	r.Run(":9999")
}
