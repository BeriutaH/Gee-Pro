package main

import (
	"fmt"
	"gee"
	"net/http"
	"time"
)

type student struct {
	Name string
	Age  int8
}

func FormatData(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	//r := gee.New() // 全局共享一个Engine
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

	//// 路由组测试
	//r.GET("/index", func(ctx *gee.Context) {
	//	ctx.HTML(http.StatusOK, "<h1>Index Page</h1>")
	//})
	//
	//v1 := r.Group("/v1")
	//{
	//	v1.GET("/", func(ctx *gee.Context) {
	//		ctx.HTML(http.StatusOK, "<h1>Hello Home</h1>")
	//	})
	//	v1.GET("/hello", func(ctx *gee.Context) {
	//		ctx.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	//	})
	//
	//}
	//
	//v2 := r.Group("/admin")
	//v2.Use(gee.Logger())
	//{
	//	v2.POST("/login", func(ctx *gee.Context) {
	//		_ = ctx.JSON(http.StatusOK, gee.H{
	//			"username": ctx.PostForm("username"),
	//			"password": ctx.PostForm("password"),
	//		})
	//	})
	//	v2.POST("/info/:name", func(ctx *gee.Context) {
	//		_ = ctx.JSON(http.StatusOK, gee.H{
	//			"username": ctx.Param("name"),
	//		})
	//	})
	//}

	//// 模板测试
	//r.Use(gee.Logger())
	//r.SetFuncMap(template.FuncMap{"FormatData": FormatData})
	//r.LoadHTMLGlob("templates/*")
	//r.Static("/assets", "./static")
	//stu1 := &student{Name: "Beriuta", Age: 40}
	//stu2 := &student{Name: "Jack", Age: 30}
	//r.GET("/", func(ctx *gee.Context) {
	//	ctx.HTML(http.StatusOK, "gee.tmpl", nil)
	//})
	//r.GET("/students", func(ctx *gee.Context) {
	//	ctx.HTML(http.StatusOK, "arr.tmpl", gee.H{"title": "gee", "stuArr": [2]*student{stu1, stu2}})
	//})
	//r.GET("/data", func(ctx *gee.Context) {
	//	ctx.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
	//		"title": "gee",
	//		"now":   time.Date(2024, 6, 17, 13, 23, 48, 0, time.UTC),
	//	})
	//})

	// 错误测试
	r := gee.Default()
	r.GET("/", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "Hello Gee\n")
	})
	r.GET("/panic", func(ctx *gee.Context) {
		names := []string{"gee"}
		ctx.String(http.StatusOK, names[234])
	})
	r.Run(":9999")
}
