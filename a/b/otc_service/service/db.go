package service

import (
	"time"
	"strings"

	"galaxyotc/common/utils"

	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
	"github.com/jinzhu/gorm"
	"galaxyotc/common/model"
	"github.com/spf13/viper"
	"galaxyotc/common/log"
	"galaxyotc/gc_services/otc_service/api"
)

type db struct {
	*gorm.DB
}

func newDB() *db {
	return &db{model.DB}
}

const (
	WaitingApproved = 1
	WaitingPay = 2
)

// 订单超时取消回调函数
func (db *db) OrderTimeoutCancelCallback (sn string, scene int32) (bool, error) {
	var order model.Order

	// 获取订单信息
	if err := db.First(&order, model.Order{Sn: sn}).Error; err != nil {
		log.Errorf("db-OrderTimeoutCancelCallback-Error: %s", err.Error())
		return false, err
	}

	// 订单取消场景，是超时未接单还是超时未付款
	if scene == WaitingApproved {
		cancelReason := "超时未接单"
		// 订单状态还处于接单中，则取消订单
		if order.Status == model.OrderWaitingInt {
			if err := db.Model(&order).Updates(model.Order{Status: model.OrderCanceledInt, CancelReason: cancelReason}).Error; err != nil {
				log.Errorf("db-OrderTimeoutCancelCallback-Error: %s", err.Error())
				return false, err
			}
		}
	} else if scene == WaitingPay {
		cancelReason := "超时未付款"
		// 订单状态还处于等待付款中，则取消订单
		if order.Status == model.OrderWaitingPayInt {
			if err := db.Model(&order).Updates(model.Order{Status: model.OrderCanceledInt, CancelReason: cancelReason}).Error; err != nil {
				log.Errorf("db-OrderTimeoutCancelCallback-Error: %s", err.Error())
				return false, err
			}
		}
	}

	return true, nil
}

// 订单超时放币回调函数
func (db *db) OrderTimeoutReleaseCallback (sn string) (bool, error) {
	var order model.Order

	// 获取订单信息
	if err := db.First(&order, model.Order{Sn: sn}).Error; err != nil {
		log.Errorf("db-OrderTimeoutReleaseCallback-Error: %s", err.Error())
		return false, err
	}

	// 订单状态为待放币状态且交易哈希为空
	if order.Status == model.OrderWaitingReleaseInt && order.Txid == "" {
		// 获取代币信息
		currency, err := model.CurrencyFromAndToRedis(order.Currency)
		if err != nil {
			log.Errorf("db-OrderTimeoutReleaseCallback-Error: %s", err.Error())
			return false, err
		}

		// 将小数转换为位
		amountWei := utils.ToWei(order.Amount, int(currency.Precision))
		// 将订单流水号的前缀去掉
		sn := strings.TrimLeft(order.Sn, viper.GetString("server.order_prefix"))

		txid, err := api.PrivateWalletApi.ExecuteTransaction(amountWei.String(), sn, 2, 0, order.BuyerAddress, order.SellerAddress, currency.PrivateTokenAddress, privateWalletApi.ToBuyer)
		if err != nil {
			log.Errorf("db-OrderTimeoutReleaseCallback-Error: %s", err.Error())
			return false, err
		}

		doneAt := time.Now().Local()
		if err := db.Model(&order).Updates(model.Order{Txid: txid, DoneAt: &doneAt}).Error; err != nil {
			log.Errorf("db-OrderTimeoutReleaseCallback-Error: %s", err.Error())
			return false, err
		}
	}
	return true, nil
}

// 订单超时完成回调函数
func (db *db) OrderTimeoutCompletedCallback (sn string) (bool, error) {
	var order model.Order

	// 获取订单信息
	if err := db.First(&order, model.Order{Sn: sn}).Error; err != nil {
		log.Errorf("db-OrderTimeoutCompletedCallback-Error: %s", err.Error())
		return false, err
	}

	// 订单状态为等待确认
	if order.Status == model.OrderWaitingCompleteInt {
		// 更改状态为交易完成
		doneAt := time.Now().Local()
		if err := db.Model(&order).Updates(model.Order{Status: model.OrderCompletedInt, DoneAt: &doneAt}).Error; err != nil {
			log.Errorf("db-OrderTimeoutCompletedCallback-Error: %s", err.Error())
			return false, err
		}
	}

	return true, nil
}