package models

import "github.com/astaxie/beego/orm"
import (
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type User struct { //1.首字母大写
	Id       int
	Name     string
	Password string
	Articles []*Article `orm:"rel(m2m)"`
}

type Article struct {
	Id2         int          `orm:"pk;auto"`
	Title       string       `orm:"size(20)"`                    //文章标题
	Content     string       `orm:"size(500)"`                   //内容
	Img         string       `orm:"size(50);null"`               //图片（路径）
	Time        time.Time    `orm:"type(datetime);auto_now_add"` //发布时间
	Count       int          `orm:"default(0)"`                  //阅读量
	Articletype *Articletype `orm:"rel(fk)"`
	Users       []*User      `orm:"reverse(many)"`
}

//文章内型表

type Articletype struct {
	Id              int
	Articletypename string     `orm:"size(500)"`
	Articles        []*Article `orm:"reverse(many)"`
}

func init() {
	//2这个注册不需要关闭
	//注册数据库
	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/classone?charset=utf8&loc=Asia%2FShanghai")

	//注册表
	orm.RegisterModel(new(User), new(Article), new(Articletype))

	//生成表  强制更新 是否看操作过程
	orm.RunSyncdb("default", false, true)

}
