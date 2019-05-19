package menuProduct

import (
	"net/http"
	"galaxyotc/common/model"
	"github.com/gin-gonic/gin"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
)

type ProductList []*model.OperateProduct

type MenuLists []*model.OperateMenu


//获取菜单
func MenuList(c *gin.Context) {

	menuLists:=MenuLists{}
	model.DB.Model(model.OperateMenu{}).Where("status = 0").Find(&menuLists).RecordNotFound();
	for i := 0; i < len(menuLists)-1; i++ {
		for j := 0; j < len(menuLists)-1-i; j++ {
			if  menuLists[j].Sort>menuLists[j+1].Sort{
				menuLists[j+1],menuLists[j] = menuLists[j],menuLists[j+1]

			}
		}
	}


	c.JSON(http.StatusOK, gin.H {
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H {
			"menus": menuLists,
		},
	})
}



// 获取菜单产品
func MenuProduct(c *gin.Context) {
	menuProductID := c.Query("menu_id")
	menuProductID=" menu_id = " +menuProductID+" and status = 0"

	menuProductIDZero:=menuProductID +" and category_id = 0 "
	menuProductZero:=ProductList{}

	menuProductIDOne:=menuProductID+" and category_id = 1"
	menuProductOne:=ProductList{}

	menuProductIDTwo:=menuProductID+" and category_id = 2"
	menuProductTwo:=ProductList{}
  

	model.DB.Model(model.OperateProduct{}).Where(menuProductIDZero).Find(&menuProductZero);
	for i := 0; i < len(menuProductZero)-1; i++ {
		for j := 0; j < len(menuProductZero)-1-i; j++ {
			if  menuProductZero[j].Sort>menuProductZero[j+1].Sort{
				menuProductZero[j+1],menuProductZero[j] = menuProductZero[j],menuProductZero[j+1]

			}
		}
	}


	model.DB.Model(model.OperateProduct{}).Where(menuProductIDOne).Find(&menuProductOne);
	for i := 0; i < len(menuProductOne)-1; i++ {
		for j := 0; j < len(menuProductOne)-1-i; j++ {
			if  menuProductOne[j].Sort>menuProductOne[j+1].Sort{
				menuProductOne[j+1],menuProductOne[j] = menuProductOne[j],menuProductOne[j+1]

			}
		}
	}

	model.DB.Model(model.OperateProduct{}).Where(menuProductIDTwo).Find(&menuProductTwo);
	for i := 0; i < len(menuProductTwo)-1; i++ {
		for j := 0; j < len(menuProductTwo)-1-i; j++ {
			if  menuProductTwo[j].Sort>menuProductTwo[j+1].Sort{
				menuProductTwo[j+1],menuProductTwo[j] = menuProductTwo[j],menuProductTwo[j+1]

			}
		}
	}

	//for i := 0; i <= len(menuProduct)-1; i++ {
	//	log.Info( menuProduct[i].Url)
	//}

	c.JSON(http.StatusOK, gin.H {
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H {
			"producsZero": menuProductZero,
			"producsOne": menuProductOne,
			"producsTwo": menuProductTwo,
			},
	})
}




func FindmenuProduct(c *gin.Context)  {
	SendErrJSON := net.SendErrJSON

	// 获取页数和条数
	//page, size := net.GetPageAndSize(c)
	//
	//// 计算起始位置
	//offset := (page - 1) * size


	name:= c.Query("name")

	nameString:="%"+name+"%"

	operateProductList:=[]model.OperateProduct{}

	baseQuery := model.DB.Model(&model.OperateProduct{}).Where(" name like ?",nameString)


	//if err := baseQuery.Offset(offset).Limit(size).Find(&operateProductList).Error; err != nil {
	//	log.Errorf("User-Signin-Error: %s", err.Error())
	//	SendErrJSON("查询错误", c)
	//	return
	//}
	if err := baseQuery.Find(&operateProductList).Error; err != nil {
		log.Errorf("User-Signin-Error: %s", err.Error())
		SendErrJSON("查询错误", c)
		return
	}

	c.JSON(http.StatusOK, gin.H {
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H {
			"OperateProductList": operateProductList,
		},
	})
}