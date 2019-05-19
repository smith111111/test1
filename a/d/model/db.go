package model

import (
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/globalsign/mgo"
	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"galaxyotc/common/log"
)

// DB 数据库连接
var DB *gorm.DB

// RedisPool Redis连接池
var RedisPool *redis.Pool

// MongoDB 数据库连接
var MongoDB *mgo.Database

func NewDB(dsn string) {
	db, err := gorm.Open(viper.GetString("db.dialect"), dsn)
	if err != nil {
		log.Fatalf("Open DB Error: %s", err.Error())
		os.Exit(-1)
	}
	if viper.GetBool("server.dev") {
		db.LogMode(true)
	}
	db.DB().SetMaxIdleConns(viper.GetInt("db.max_idle"))
	db.DB().SetMaxOpenConns(viper.GetInt("db.max_open"))
	db.AutoMigrate(
		&User{},
		&Deposit{},
		&Withdraw{},
		&Currency{},
		&FiatCurrency{},
		&TradingMethod{},
		&Offer{},
		&Order{},
		&Area{},
		&RealnameVerification{},
		&UserTradingMethod{},
		&TradingRate{},
		&Notice{},
		&CommissionDistribution{},
		&CommissionDistributionReceipt{},
		&ImMsg{},
		&PushMsg{},
		&PushToken{},
		&PushAppKey{},
		&OperateMenu{},
		&OperateProduct{},
		&Transfer{},
		&AppVersion{},
		&Financial{},
		&FinancialEarningsHis{},
		&FinancialOrder{},
		&FinancialProduct{},
	)
	DB = db
}

func NewRedis() {
	RedisPool = &redis.Pool{
		MaxIdle:     viper.GetInt("redis.max_idle"),
		MaxActive:   viper.GetInt("redis.max_active"),
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			var options []redis.DialOption
			options = append(options, redis.DialPassword(viper.GetString("redis.password")), redis.DialDatabase(viper.GetInt("redis.data_base")))
			c, err := redis.Dial("tcp", viper.GetString("redis.url"), options...)
			if err != nil {
				log.Fatalf("Open Redis Error: %s", err.Error())
				return nil, err
			}
			return c, nil
		},
	}
}

/*
 * mgo文档 http://labix.org/mgo
 * https://godoc.org/gopkg.in/mgo.v2
 * https://godoc.org/gopkg.in/mgo.v2/bson
 * https://godoc.org/gopkg.in/mgo.v2/txn
 */
func NewMongo() {
	MongoUrl := viper.GetString("mongo.url")
	if MongoUrl == "" {
		return
	}
	session, err := mgo.Dial(MongoUrl)
	if err != nil {
		log.Fatalf("Open MongoDB Error: %s", err.Error())
		os.Exit(-1)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	MongoDB = session.DB(viper.GetString("mongo.data_base"))
}