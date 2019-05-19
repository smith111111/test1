package exchangerates

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/op/go-logging"
	"golang.org/x/net/proxy"
	"strings"
)

const SatoshiPerBTC = 100000000

var log = logging.MustGetLogger("exchangeRates")

// 汇率提供者：即，一个数字货币交易所。
type ExchangeRateProvider struct {
	// 数据抓取URL地址
	fetchUrl string
	// 汇率列表缓存：币种 - 1个BTC的金额
	cache    map[string]float64
	// HTTP客户端
	client   *http.Client
	// 汇率解码器
	decoder  ExchangeRateDecoder
}

// 汇率解码器接口
type ExchangeRateDecoder interface {
	// 解码： JSON对象， 汇率表缓存
	decode(dat interface{}, cache map[string]float64) (err error)
}

// 不同的汇率解码器实现
// empty structs to tag the different ExchangeRateDecoder implementations
type BitcoinAverageDecoder struct{}
type BitPayDecoder struct{}
type BlockchainInfoDecoder struct{}
type BitcoinChartsDecoder struct{}

// 比特币价格获取器
type BitcoinPriceFetcher struct {
	// 并发控制锁
	sync.Mutex
	// 汇率列表缓存：币种 - 1个BTC的金额
	cache     map[string]float64
	// 提供者列表
	providers []*ExchangeRateProvider
}

// 构建一个比特币价格获取器实例：Tor代理
func NewBitcoinPriceFetcher(dialer proxy.Dialer) *BitcoinPriceFetcher {
	b := BitcoinPriceFetcher{
		cache: make(map[string]float64),
	}
	dial := net.Dial
	if dialer != nil { // 如果提供了一个代理
		dial = dialer.Dial
	}
	tbTransport := &http.Transport{Dial: dial}
	// 构建一个HTTP客户端实例
	client := &http.Client{Transport: tbTransport, Timeout: time.Minute}

	// 4个汇率提供者：实例化时，将b.cache缓存传递给了它们
	b.providers = []*ExchangeRateProvider{
		{"https://ticker.openbazaar.org/api", b.cache, client, BitcoinAverageDecoder{}},
		{"https://bitpay.com/api/rates", b.cache, client, BitPayDecoder{}},
		{"https://blockchain.info/ticker", b.cache, client, BlockchainInfoDecoder{}},
		{"https://api.bitcoincharts.com/v1/weighted_prices.json", b.cache, client, BitcoinChartsDecoder{}},
	}
	// 运行一个例程，每5钟抓取一次最新的行情数据
	go b.run()
	return &b
}

// 从缓存中获取指定币种的汇率（即，1个BTC的价格）
func (b *BitcoinPriceFetcher) GetExchangeRate(currencyCode string) (float64, error) {
	// 币种转为大写
	currencyCode = NormalizeCurrencyCode(currencyCode)

	// 请求一个锁
	b.Lock()
	// 函数结束时，释放一个锁
	defer b.Unlock()
	// 从缓存中获取指定币种的汇率（即，1个BTC的价格）
	price, ok := b.cache[currencyCode]
	if !ok {
		return 0, errors.New("Currency not tracked")
	}
	return price, nil
}

// 获取最新指定币种的汇率（即，1个BTC的价格），并更新缓存。
func (b *BitcoinPriceFetcher) GetLatestRate(currencyCode string) (float64, error) {
	currencyCode = NormalizeCurrencyCode(currencyCode)

	b.fetchCurrentRates()
	b.Lock()
	defer b.Unlock()
	price, ok := b.cache[currencyCode]
	if !ok {
		return 0, errors.New("Currency not tracked")
	}
	return price, nil
}

// 获取汇率表：是否从缓存中获取？
func (b *BitcoinPriceFetcher) GetAllRates(cacheOK bool) (map[string]float64, error) {
	if !cacheOK {
		// 从提供者哪里抓取当前汇率信息，并更新缓存信息
		err := b.fetchCurrentRates()
		if err != nil {
			return nil, err
		}
	}
	// 计算GC币的价值
	GC := b.cache["CNY"] / 1.5
	b.cache["GC"] = GC

	b.Lock()
	defer b.Unlock()
	return b.cache, nil
}

// 返回1个BTC对应的最小单位金额：1个BTC= 10^9个中本聪
func (b *BitcoinPriceFetcher) UnitsPerCoin() int {
	return SatoshiPerBTC
}

// 从提供者哪里抓取当前汇率信息，并更新缓存信息
func (b *BitcoinPriceFetcher) fetchCurrentRates() error {

	b.Lock()
	defer b.Unlock()
	// 遍历每个提供者
	for _, provider := range b.providers {
		// 从提供者那里获取汇率信息，并更新自身缓存。
		err := provider.fetch()
		if err == nil {
			return nil
		}
	}
	log.Error("Failed to fetch bitcoin exchange rates")
	return errors.New("All exchange rate API queries failed")
}

// 从提供者那里获取汇率信息，并更新自身缓存。
func (provider *ExchangeRateProvider) fetch() (err error) {
	if len(provider.fetchUrl) == 0 {
		err = errors.New("Provider has no fetchUrl")
		return err
	}
	// 抓取数据
	resp, err := provider.client.Get(provider.fetchUrl)
	if err != nil {
		log.Error("Failed to fetch from "+provider.fetchUrl, err)
		return err
	}
	// 将响应内容解码成一个JSON对象（即一个interface{}实例）
	decoder := json.NewDecoder(resp.Body)
	var dataMap interface{}
	err = decoder.Decode(&dataMap)
	if err != nil {
		log.Error("Failed to decode JSON from "+provider.fetchUrl, err)
		return err
	}
	// 解码JSON对象来更新自身汇率缓存信息
	return provider.decoder.decode(dataMap, provider.cache)
}

// 将在一个例程中运行，每5钟抓取一次最新的行情数据
func (b *BitcoinPriceFetcher) run() {
	// 从提供者哪里抓取当前汇率信息，并更新缓存信息
	b.fetchCurrentRates()
	ticker := time.NewTicker(time.Minute * 5)
	for range ticker.C { // 通道阻塞5分钟等待一个数据
		// 从提供者哪里抓取当前汇率信息，并更新缓存信息
		b.fetchCurrentRates()
	}
}

// Decoders
func (b BitcoinAverageDecoder) decode(dat interface{}, cache map[string]float64) (err error) {
	// JSON对象转换成一个字典：币种 - 行情JSON对象
	data := dat.(map[string]interface{})
	for k, v := range data {
		if k != "timestamp" {
			// 行情JSON对象 转换成一个字典： 属性 - 值
			val, ok := v.(map[string]interface{})
			if !ok {
				return errors.New(reflect.TypeOf(b).Name() + ".decode: Type assertion failed")
			}
			price, ok := val["last"].(float64) // 最新价格
			if !ok {
				return errors.New(reflect.TypeOf(b).Name() + ".decode: Type assertion failed, missing 'last' (float) field")
			}
			cache[k] = price
		}
	}
	return nil
}

func (b BitPayDecoder) decode(dat interface{}, cache map[string]float64) (err error) {
	// JSON对象转换成一个数组： 行情信息JSON对象列表
	data := dat.([]interface{})
	for _, obj := range data {
		// JSON对象转换成一个字典：属性 - 值
		code := obj.(map[string]interface{})
		k, ok := code["code"].(string) // 币种
		if !ok {
			return errors.New(reflect.TypeOf(b).Name() + ".decode: Type assertion failed, missing 'code' (string) field")
		}
		price, ok := code["rate"].(float64) // 汇率
		if !ok {
			return errors.New(reflect.TypeOf(b).Name() + ".decode: Type assertion failed, missing 'rate' (float) field")
		}
		cache[k] = price
	}
	return nil
}

func (b BlockchainInfoDecoder) decode(dat interface{}, cache map[string]float64) (err error) {
	// JSON对象转换成一个字典：币种 - 行情JSON对象
	data := dat.(map[string]interface{})
	for k, v := range data {
		// 行情JSON对象转换成字典：属性 - 值
		val, ok := v.(map[string]interface{})
		if !ok {
			return errors.New(reflect.TypeOf(b).Name() + ".decode: Type assertion failed")
		}
		price, ok := val["last"].(float64) // 价格
		if !ok {
			return errors.New(reflect.TypeOf(b).Name() + ".decode: Type assertion failed, missing 'last' (float) field")
		}
		cache[k] = price
	}
	return nil
}

func (b BitcoinChartsDecoder) decode(dat interface{}, cache map[string]float64) (err error) {
	// JSON对象转换成一个字典：币种 - 行情JSON对象
	data := dat.(map[string]interface{})
	for k, v := range data {
		if k != "timestamp" {
			// 行情JSON对象转换成字典：属性 - 值
			val, ok := v.(map[string]interface{})
			if !ok {
				return errors.New("Type assertion failed")
			}
			p, ok := val["24h"] // 24小时行情价格
			if !ok {
				continue
			}
			pr, ok := p.(string)
			if !ok {
				return errors.New("Type assertion failed")
			}
			price, err := strconv.ParseFloat(pr, 64)
			if err != nil {
				return err
			}
			cache[k] = price
		}
	}
	return nil
}

// NormalizeCurrencyCode standardizes the format for the given currency code
func NormalizeCurrencyCode(currencyCode string) string {
	return strings.ToUpper(currencyCode)
}
