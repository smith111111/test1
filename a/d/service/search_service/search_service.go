package search_service

import (
	"context"
	"fmt"
	"github.com/olivere/elastic"
	"galaxyotc/common/log"
	"sync"
	"reflect"
	"github.com/pkg/errors"
	"github.com/panjf2000/ants"
)

var (
	_galaxyOtcIndex string = "galaxyotc"
)

const (
	TYPE_USER  = "user"
	TYPE_OFFER = "offer"
)

type SearchUserInfo struct {
	Id     uint64 `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
	Email  string `json:"email"`
}

type SearchOfferInfo struct {
	Id               uint64   `json:"id"`
	OfferType        int32    `json:"offer_type"`         	// 广告类型 出售或购买
	CurrencyCode     string `json:"currency_code"`      	// 代币代码
	FiatCurrency     uint   `json:"fiat_currency"`      	// 法币ID
	FiatCurrencyCode string `json:"fiat_currency_code"` 	// 法币代码
	PublisherId      uint64   `json:"publisher_id"`       	// 发布者ID
	PublisherName    string `json:"publisher_name"`     	// 发布者名称
	OfferStatus      int    `json:"offer_status"`       	// 广告状态
}

var client *Client

type Client struct {
	es  *elastic.Client
	ctx context.Context
	ap 	*ants.Pool
}

// 创建新连接
func NewClient(addrs []string) *Client {
	var once sync.Once
	once.Do(func() {
		es, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(addrs...))
		if err != nil {
			log.Errorf("api-NewClient-error: %s", err.Error())
		}
		ap, err := ants.NewPool(50)
		if err != nil {
			log.Errorf("api-NewClient-error: %s", err.Error())
		}
		client = &Client{
			es: es,
			ctx: context.Background(),
			ap: ap,
		}
	})
	return client
}

func (c *Client) Start() {
	client._initMapping(_galaxyOtcIndex)
}

func (c *Client) _initMapping(esIndexName string) error {
	var err error
	exists, err := c.es.IndexExists(esIndexName).Do(c.ctx)
	if err != nil {
		log.Error("IndexExists" + err.Error())
		return err
	}

	if !exists {
		c.createUserMapping(esIndexName)
		c.createOfferMapping(esIndexName)
	}

	return err
}

func (c *Client) createUserMapping(esIndexName string) error {
	indexMapping := `{"mappings":{"user":{"properties":{"name":{"analyzer":"ik_max_word","type":"text"},"mobile":{"analyzer":"ik_max_word","type":"text"},"email":{"analyzer":"ik_max_word","type":"text"}}}}}`
	createIndex, err := c.es.CreateIndex(esIndexName).Body(indexMapping).Do(c.ctx)
	if err != nil {
		log.Errorf("api-createUserMapping-error: %s", err.Error())
		return err
	}

	if !createIndex.Acknowledged {
		return errors.New("create index:" + esIndexName + ", not Ack nowledged")
	}

	return nil
}

func (c *Client) createOfferMapping(esIndexName string) error {
	indexMapping := `{"mappings":{"offer":{"properties":{"publisher_name":{"analyzer":"ik_max_word","type":"text"}}}}}`
	createIndex, err := c.es.CreateIndex(esIndexName).Body(indexMapping).Do(c.ctx)
	if err != nil {
		log.Errorf("api-createOfferMapping-error: %s", err.Error())
		return err
	}

	if !createIndex.Acknowledged {
		return errors.New("create index:" + esIndexName + ", not Ack nowledged")
	}

	return nil
}

func (c *Client) AddUserInfo(uid uint64, data *SearchUserInfo) error {
	_, err := c.es.Index().Index(_galaxyOtcIndex).Type(TYPE_USER).Id(fmt.Sprintf("%d", uid)).BodyJson(data).Do(c.ctx)
	if err != nil {
		log.Errorf("api-AddUserInfo-error: %s", err.Error())
		return err
	}

	return nil
}

func (c *Client) AsyncAddUserInfo(uid uint64, data *SearchUserInfo) {
	c.ap.Submit(func() error {
		err := c.AddUserInfo(uid, data)
		if err != nil {
			log.Errorf("api-AsyncAddUserInfo-error: %s", err.Error())
			return nil
		}

		return nil
	})
}

func (c *Client) UpdateUserInfo(uid uint64, data map[string]interface{}) error {
	_, err := c.es.Update().Index(_galaxyOtcIndex).Type(TYPE_USER).Id(fmt.Sprintf("%d", uid)).Doc(data).Do(c.ctx)
	if err != nil {
		log.Errorf("api-UpdateUserInfo-error: %s", err.Error())
		return err
	}

	return nil
}

func (c *Client) AsyncUpdateUserInfo(uid uint64, data map[string]interface{}) {
	c.ap.Submit(func() error {
		err := c.UpdateUserInfo(uid, data)
		if err != nil {
			log.Errorf("api-AsyncUpdateUserInfo-error: %s", err.Error())
			return nil
		}

		return nil
	})
}

func (c *Client) SearchUserInfo(keyword string, from, size int) (list []*SearchUserInfo, err error) {
	list = make([]*SearchUserInfo, 0)

	boolQ := elastic.NewBoolQuery()
	boolQ.Should(elastic.NewMatchQuery("name", keyword))
	searchResult, err := c.es.Search().Index(_galaxyOtcIndex).Type(TYPE_USER).Query(boolQ).From(from).Size(size).Do(c.ctx)
	if err != nil {
		log.Errorf("api-SearchUserInfo-error: %s", err.Error())
		return list, err
	}

	for _, item := range searchResult.Each(reflect.TypeOf(&SearchUserInfo{})) {
		list = append(list, item.(*SearchUserInfo))
	}

	return list, nil
}

func (c *Client) AddOfferInfo(offerId uint64, data *SearchOfferInfo) error {
	_, err := c.es.Index().Index(_galaxyOtcIndex).Type(TYPE_OFFER).Id(fmt.Sprintf("%d", offerId)).BodyJson(data).Do(c.ctx)
	if err != nil {
		log.Errorf("api-AddOfferInfo-error: %s", err.Error())
		return err
	}

	return nil
}

func (c *Client) AsyncAddOfferInfo(offerId uint64, data *SearchOfferInfo) {
	c.ap.Submit(func() error {
		err := c.AddOfferInfo(offerId, data)
		if err != nil {
			log.Errorf("api-AsyncAddOfferInfo-error: %s", err.Error())
			return nil
		}

		return nil
	})
}

func (c *Client) UpdateOfferInfo(offerId uint64, data map[string]interface{}) error {
	_, err := c.es.Update().Index(_galaxyOtcIndex).Type(TYPE_OFFER).Id(fmt.Sprintf("%d", offerId)).Doc(data).Do(c.ctx)
	if err != nil {
		log.Errorf("api-UpdateOfferInfo-error: %s", err.Error())
		return err
	}

	return nil
}

func (c *Client) AsyncUpdateOfferInfo(offerId uint64, data map[string]interface{}) {
	c.ap.Submit(func() error {
		err := c.UpdateOfferInfo(offerId, data)
		if err != nil {
			log.Errorf("api-AsyncUpdateOfferInfo-error: %s", err.Error())
			return nil
		}

		return nil
	})
}

func (c *Client) SearchOfferInfo(keyword string, offerType int, currencyCode string, fiatCurrencyCode string, from, size int) (list []*SearchOfferInfo, err error) {
	list = make([]*SearchOfferInfo, 0)

	boolQ := elastic.NewBoolQuery()
	boolQ.Should(elastic.NewMatchQuery("publisher_name", keyword), elastic.NewMatchQuery("offer_type", offerType), elastic.NewMatchQuery("currency_code", currencyCode), elastic.NewMatchQuery("fiat_currency_code", fiatCurrencyCode))
	searchResult, err := c.es.Search().Index(_galaxyOtcIndex).Type(TYPE_OFFER).Query(boolQ).From(from).Size(size).Do(c.ctx)
	if err != nil {
		log.Errorf("api-SearchOfferInfo-error: %s", err.Error())
		return list, err
	}

	for _, item := range searchResult.Each(reflect.TypeOf(&SearchOfferInfo{})) {
		list = append(list, item.(*SearchOfferInfo))
	}

	return list, nil
}

func (c *Client) DelOfferInfo(offerId uint) error {
	_, err := c.es.Delete().Index(_galaxyOtcIndex).Type(TYPE_OFFER).Id(fmt.Sprintf("%d", offerId)).Do(c.ctx)
	if err != nil {
		log.Errorf("api-DelOfferInfo-error: %s", err.Error())
		return err
	}

	return nil
}
