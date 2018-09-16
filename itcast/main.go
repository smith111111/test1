package main

import (
	"github.com/astaxie/beego"
	_ "itcast/models"
	_ "itcast/routers"
	"strconv"
)

//初始化连接
func HandlerBeforeIndex(data int) string {
	beego.Info(data)
	pageIndex := data - 1
	pageIndex1 := strconv.Itoa(pageIndex)
	return pageIndex1
}

func HandlerAfterIndex(data int) int {
	pageIndex := data + 1

	return pageIndex
}

func main() {
	beego.AddFuncMap("BeforeIndex", HandlerBeforeIndex)
	beego.AddFuncMap("AfterIndex", HandlerAfterIndex)
	beego.Run()

}
