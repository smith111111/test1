package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"itcast/models"
	"path"
	"time"
	"math"
	"fmt"
)

type ArticleController struct {
	beego.Controller
}

func (this *ArticleController) ShowArticle() {
	//重点get 这里是nil
	//name:=this.GetSession("userName")
	//if name==nil{
	//	beego.Info("用户没登陆")
	//	this.Redirect("/",302)
	//	return
	//}
	selectOpt := this.GetString("select")
	//获取页面
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}

	//每页长度
	pageSize := 2
	start := (pageIndex - 1) * pageSize

	//回到汇总页面,高级查询
	orm := orm.NewOrm()
	var articles []models.Article
	var typeAc []models.Articletype
	orm.QueryTable("articletype").All(&typeAc)
	qs := orm.QueryTable("article")

	var count int64
	if selectOpt == "" { //直接进去
		count, _ = qs.Count()
	} else {
		//查询总数
		count, _ = qs.RelatedSel("Articletype").Filter("Articletype__Articletypename", selectOpt).Count() //返回数据条目数   加过滤器
	}
	countPage1:=float64(count)/float64(pageSize)
	//总页数
	 countPage:= math.Ceil(countPage1)

	//	qs.Limit(pageSize, start).All(&articles)    RelatedSel("Articletype").All(&articles)

	//beego.Info(articles[0].Articletype.Articletypename) selectOpt

	if selectOpt == "" { //直接进去
		qs.Limit(pageSize, start).RelatedSel("Articletype").All(&articles)
	} else {
		qs.Limit(pageSize, start).RelatedSel("Articletype").
			Filter("Articletype__Articletypename", selectOpt).All(&articles)
		//	All(&articles)
	}

	fristPage := false
	endPage := false
	if pageIndex == 1 {
		fristPage = true
	}
	if pageIndex == int(countPage) {
		endPage = true
	}

	this.Data["articles"] = articles
	this.Data["articleCount"] = count
	this.Data["countPage"] = countPage
	this.Data["pageIndex"] = pageIndex

	this.Data["fristPage"] = fristPage
	this.Data["endPage"] = endPage
	this.Data["typeAc"] = typeAc
	this.Data["selectOpt"]=selectOpt

	//articleTypesi

	this.Layout = "layout.html"
	this.LayoutSections=make(map[string]string)
	this.Data["layoutTile"]="首页"
	userName:=this.GetSession("userName");
	this.Data["username"]=userName
	//渲染
	this.TplName = "index.html"

}

func (this *ArticleController) ShowAddArticle() {
	orm := orm.NewOrm()
	var articleTypes []models.Articletype
	orm.QueryTable("Articletype").All(&articleTypes)
	this.Data["articleTypes"] = articleTypes
	this.Layout = "layout.html"
	this.LayoutSections=make(map[string]string)
//	this.LayoutSections["layoutTile"]="<title>添加类型</title>"
	this.Data["layoutTile"]="添加类型"
	this.TplName = "add.html"
}

func (this *ArticleController) HandlerArticlePost() {

	//获取数据，处理数据，插入数据，返回页面

	articleName := this.GetString("articleName")

	content := this.GetString("content")
	oSelect := this.GetString("select")

	beego.Info(oSelect)
	//uploadname:=this.GetString("uploadname")
	//
	f, s, errs := this.GetFile("uploadname") //2

	defer f.Close()

	endName := path.Ext(s.Filename) //1
	//这里本来是数组循环遍历
	if endName != ".png" && endName != ".jpg" {
		beego.Info("文件格式错误")
		this.Redirect("/article", 302)
		return
	}
	if s.Size > 5000000 {
		beego.Info("文件太大")
		this.Redirect("/article", 302)
		return
	}
	beego.Info(time.Now(), "现在")
	//fileName:=time.Now().Format("2006-01-02 15:04:05")
	fileName := time.Now().Format("2006-01-02 15:04:05")

	err := this.SaveToFile("uploadname", "./static/img/"+fileName+endName) //3
	if err != nil {
		beego.Info("上传错误")
		this.Redirect("/article", 302)
		return
	}
	if errs != nil {
		beego.Info("文件错误")
		this.Redirect("/article", 302)
		return
	}
	//	beego.Info(articleName) 每有类型插入不了
	orm := orm.NewOrm()
	articletype := models.Articletype{}
	articletype.Articletypename = oSelect
	orm.Read(&articletype, "Articletypename")

	art := models.Article{}
	art.Img = "./static/img/" + fileName + endName
	art.Title = articleName
	art.Content = content
	art.Articletype = &articletype

	orm.Insert(&art) //插入对象

	this.Redirect("/Article/article", 302)

}

func (this *ArticleController) ShowContent() {
	id, _ := this.GetInt("id2")
	orm := orm.NewOrm()
	article := models.Article{}
	article.Id2 = id
	orm.Read(&article)

	this.Data["article"] = article

	//更新数据数据
	article.Count += 1

	orm.Update(&article) //默认当前用户

	//插入数据 和查询数据去重
	userName:=this.GetSession("userName")
	user:=models.User{}
	user.Name=userName.(string)
	orm.Read(&user,"name")
	m2m:=orm.QueryM2M(&article,"Users")
	_,err:=m2m.Add(&user)
	if err!=nil{
		fmt.Println("插入失败")
		return
	}

	//查文章中多少用户，先查用户

	var userArray []models.User
	orm.QueryTable("User").Filter("Articles__Article__Id2",id).Distinct().All(&userArray)

	this.Data["userArray"]=userArray
	this.TplName = "content.html"
}

//todo回主面

func (this *ArticleController) HandlerContentPost() {

}

func (this *ArticleController) ShowDelete() {
	id, _ := this.GetInt("id2")
	orm := orm.NewOrm()
	article := models.Article{}
	article.Id2 = id
	orm.Delete(&article)
	this.Redirect("/Article/article", 302)
}

//todo回主面
func (this *ArticleController) HandlerDeletePost() {

}

func (this *ArticleController) ShowEdit() {
	id, _ := this.GetInt("id2")
	orm := orm.NewOrm()
	article := models.Article{}
	article.Id2 = id
	orm.Read(&article)
	this.Data["article"] = article
	this.TplName = "update.html"
	//this.Redirect("/article",302)
}

//todo回主面
func (this *ArticleController) HandlerEditPost() {
	id, _ := this.GetInt("id2")
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	f, s, errs := this.GetFile("uploadname") //2
	defer f.Close()

	endName := path.Ext(s.Filename) //1

	orm := orm.NewOrm()

	article := models.Article{Id2: id}

	err := orm.Read(&article)
	if err != nil {
		beego.Info("查询数据错误", err)
		return
	}

	article.Title = articleName
	article.Content = content
	if endName != "" {
		//这里本来是数组循环遍历
		if endName != ".png" && endName != ".jpg" {
			beego.Info("文件格式错误")
			this.Redirect("/Article/article", 302)
			return
		}
		if s.Size > 5000000 {
			beego.Info("文件太大")
			this.Redirect("/Article/article", 302)
			return
		}
		//fileName:=time.Now().Format("2006-01-02 15:04:05")
		fileName := time.Now().Format("2006-01-02 15:04:05")

		err := this.SaveToFile("uploadname", "./static/img/"+fileName+endName) //3
		if err != nil {
			beego.Info("上传错误")
			this.Redirect("/Article/article", 302)
			return
		}
		if errs != nil {
			beego.Info("文件错误")
			this.Redirect("/Article/article", 302)
			return
		}
		article.Img = "./static/img/" + fileName + endName
	}
	orm.Update(&article)
	this.Redirect("article", 302)

}

func (this *ArticleController) ShowArticleType() {

	orm := orm.NewOrm()
	var articletypes []models.Articletype
	orm.QueryTable("Articletype").All(&articletypes)
	this.Data["articletypes"] = articletypes

	this.TplName = "addType.html"
}

func (this *ArticleController) HandlerArticleTypePost() {
	typeName := this.GetString("typeName")
	beego.Info(typeName)
	beego.Info(typeName)
	if typeName == "" {
		beego.Info("请输入数据")
		this.TplName = "addType.html"
		return
	}
	orm := orm.NewOrm()
	articleTypw := models.Articletype{}

	articleTypw.Articletypename = typeName
	_, err := orm.Insert(&articleTypw)
	if err != nil {
		beego.Info("插入失败", err)
		return
	}

	this.Redirect("/Article/AddArticleType", 302)
}

//删除文章类型
func (this *ArticleController)DeleteArticleType()  {
	id,_:=this.GetInt("Id")

	orm:=orm.NewOrm()
	artType :=models.Articletype{Id:id}
	orm.Delete(&artType)
	this.Redirect("/Article/AddArticleType",302)

}

func (this *ArticleController)ArticleOut()  {
	this.DelSession("userName")
	this.Redirect("/",302)
}