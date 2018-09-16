package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"itcast/models"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) ShowLogin() {
	name:=c.Ctx.GetCookie("userName")
	if name!=""{
		c.Data["userName"]= name
		c.Data["check"] = "checked"
	}
	//else {
	//	c.Data["userName"]= ""
	//	c.Data["check"] = ""
	//}
	c.TplName = "login.html"

}

func (c *MainController) HandlerLoginPost() {
	//抽取分页待处理
	name := c.GetString("userName")
	pwd := c.GetString("password")
	check := c.GetString("remember")
//cookit
	if check=="on"{
		c.Ctx.SetCookie("userName",name,3600)
	}else{
		c.Ctx.SetCookie("userName","ss",-1)
	}

	//todo后台用户名重复
	if name == "" || pwd == "" {
		beego.Info("用户名或密码为空")
		c.TplName = "login.html"
		return
	}

	orm := orm.NewOrm()
	user := models.User{}
	user.Name = name
	//****重点
	err := orm.Read(&user, "Name")
	if err != nil {
		beego.Info("用户名错误")
		c.TplName = "login.html"
		return
	}

	//加密的校验
	if pwd != user.Password {
		beego.Info("密码错误")
		c.TplName = "login.html"
		return

	}
	c.SetSession("userName",name)

	//3重点3
	c.Redirect("/Article/article", 302)

}
