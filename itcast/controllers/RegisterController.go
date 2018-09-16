package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"itcast/models"
)

type RegisterController struct {
	beego.Controller
}

func (this *RegisterController) ShowRegister() {

	this.TplName = "register.html"

}

func (this *RegisterController) HandlerPost() {

	name := this.GetString("userName")
	pwd := this.GetString("password")
	//todo后台用户名重复
	if name == "" || pwd == "" {
		beego.Info("用户名或密码为空")
		this.TplName = "register.html"
		return
	}
	//
	orm := orm.NewOrm()
	user := models.User{}
	user.Name = name
	user.Password = pwd

	_, err := orm.Insert(&user)
	if err != nil {
		beego.Info(err)
		this.TplName = "register.html"
	}
	this.Redirect("/", 302)
}
