package routers

import (
	"github.com/astaxie/beego"
	"itcast/controllers"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.Router("/", &controllers.MainController{}, "get:ShowLogin;post:HandlerLoginPost")
	//1.这里只能大写
	beego.Router("/register", &controllers.RegisterController{}, "get:ShowRegister;post:HandlerPost")

	//
	beego.Router("/Article/index", &controllers.IndexController{}, "get:ShowIndex")

	beego.Router("/Article/article", &controllers.ArticleController{}, "get:ShowArticle")

	beego.Router("/Article/addArticle", &controllers.ArticleController{}, "get:ShowAddArticle;post:HandlerArticlePost")

	beego.Router("/Article/Content", &controllers.ArticleController{}, "get:ShowContent;post:HandlerContentPost")

	beego.Router("/Article/Delete", &controllers.ArticleController{}, "get:ShowDelete;post:HandlerDeletePost")

	beego.Router("/Article/Update", &controllers.ArticleController{}, "get:ShowEdit;post:HandlerEditPost")

	beego.Router("/Article/AddArticleType", &controllers.ArticleController{}, "get:ShowArticleType;post:HandlerArticleTypePost")

	beego.InsertFilter("/Article/*",beego.BeforeRouter,FilterFunc)


	beego.Router("/Article/ArticleDeleteType", &controllers.ArticleController{}, "get:DeleteArticleType")

	beego.Router("/Article/ArticleOut", &controllers.ArticleController{}, "get:ArticleOut")

}
////注意这里最好用匿名
var FilterFunc=func(ctx *context.Context)  {
	userName:=ctx.Input.Session("userName")
	if userName ==nil{
		ctx.Redirect(302,"/")
	}



}

