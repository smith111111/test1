package controllers

import "github.com/astaxie/beego"

//

type IndexController struct {
	beego.Controller
}

func (this *IndexController) ShowIndex() {
	this.TplName = "index.html"
}
