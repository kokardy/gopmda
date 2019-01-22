package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kokardy/gopmda"
)

func main() {
	go download()
	server()
}

func download() {
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

	//静的ファイル 添付文書PDF,インタビューフォームPDF
	r.GET("/path/:path/:file", handleFile)

	//フレーム付きHTML
	r.GET("/path/:path/main", handlePath)

	//メニューフレーム
	r.GET("/path/:path/?view=toc", handleToc)

	//メインフレーム
	r.GET("/path/:path/?view=body", handleBody)

	r.Run(":8080")

}

func handleFile(c *gin.Context) {
	path := c.Param("path")
	filename := c.Param("file")
	path = filepath.Join(path, filename)
	c.File(path)
}

func handlePath(c *gin.Context) {
	path := c.Param("path")
	path = fmt.Sprintf("save/%s/", path)
	c.HTML(200, "frame.html", gin.H{
		"path": path,
	})
}
func handleToc(c *gin.Context) {
	path := c.Param("path")
	newPath := fmt.Sprintf("/path/%s/toc.html", path)
	c.Redirect(303, newPath)
}

func handleBody(c *gin.Context) {
	path := c.Param("path")
	newPath := fmt.Sprintf("/path/%s/body.html", path)
	c.Redirect(303, newPath)
}

func handleYJ(c *gin.Context) {
	yj := c.Param("yjcode")

	//YJから始まるディレクトリを探す
	dirs, err := getDirs(yj)
	if err != nil {
		c.String(404, "SERVER ERROR")
	}

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
	c.HTML(200, "choice.html", gin.H{
		"pathlist": dirs,
	})

	return
}

//YJから該当するディレクトリのスライスを生成
func getDirs(yj string) ([]string, error) {
	pattern := fmt.Sprintf("save/%s*", yj)
	result, err := filepath.Glob(pattern)
	return result, err
}
