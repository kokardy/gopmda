package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kokardy/gopmda"
)

func main() {
	//全部ダウンロード
	gopmda.DownloadAll()
	deleteList, UpdateList, err := gopmda.DeleteAndUpdate(1)
	if err != nil {
		log.Panicln("err:", err)
	}
	//削除
	deleteList.Delete()
	//追加
	UpdateList.Update()
}

func server() {
	//HTTP サーバ

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello world")
	})
	r.GET("/hoge", func(c *gin.Context) {
		c.String(200, "fuga")
	})

	//フレーム付きHTML
	r.GET("/yj/:yjcode/", handleYJ)

	//フレーム付きHTML
	r.GET("/path/:path/", handlePath)

	//メニューフレーム
	r.GET("/path/:path/?view=toc", handlePath)

	//メインフレーム
	r.GET("/path/:path/?view=body", handlePath)

	//静的ファイル 添付文書PDF,インタビューフォームPDF
	r.Static("/path", "save/")

	r.Run(":8080")

}

func handlePath(c *gin.Context) {
	path := c.Param("path")
	path = fmt.Sprintf("save/%s/", path)
	c.DinamicFile(fdsa)

}

func handleYJ(c *gin.Context) {
	yj := c.Param("yjcode")

	dirs := getDirs(yj)
	//TODO YJからフォルダ決定

	//ないとき
	if len(dirs) == 0 {
		c.String(404, "NOT FOUND")
		return
	}

	//一つに決まるとき
	if len(dirs) == 1 {
		//redirect
		newPath := fmt.Sprintf("/path/%s/", dirs[0])
		c.Redirect(303, newPath)
		return
	}

	//決まらないとき

	return
}

func getDirs(yj string) []string {
	result := make([]string, 0, 2)
	return result
}
