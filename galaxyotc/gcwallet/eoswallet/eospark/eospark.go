package eospark

import (
	"net/http"
	"sync"
	"net"
	"time"
	"fmt"
	"net/http/httputil"
	"bytes"
	"io"
	"encoding/json"
	"errors"
	"strconv"
)

type API struct {
	HttpClient 			*http.Client
	Header     			http.Header
	ApiUrl    			string
	Debug      			bool

	errorMap 			map[int]string
	apiKey				string
	lastGetInfoStamp 	time.Time
	lastGetInfoLock  	sync.Mutex
}

func New(apiURL, apiKey string) *API {
	api := &API{
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableKeepAlives:     true, // default behavior, because of `nodeos`'s lack of support for Keep alives.
			},
		},
		ApiUrl:  apiURL,
		errorMap: newEosParkErrorMap(),
		apiKey: apiKey,
		Header:   make(http.Header),
	}

	return api
}

// 查询交易信息列表
func (api *API) GetAccountRelatedTrxInfo(account string, transactionType, sort, page, size int) (out *GetAccountRelatedTrxInfoResp, err error) {
	params := make(map[string]string)
	params["module"] = "account"
	params["action"] = "get_account_related_trx_info"
	params["account"] = account

	if transactionType != 0 {
		// (1,转入 2,转出 3,全部) 默认3
		params["transaction_type"] = strconv.Itoa(transactionType)
	}
	
	if sort != 0 {
		// (1, DESC 2,ASC) 默认1
		params["sort"] = strconv.Itoa(sort)
	}
	
	if page != 0 {
		// 页码,默认第一页
		params["page"] = strconv.Itoa(page)
	}
	
	if size != 0 {
		// 每页数据条数,最大20
		params["size"] = strconv.Itoa(size)
	}
	
	err = api.call(params, &out)
	return
}

// 查询所拥有的代币列表
func (api *API) GetTokenList(account, symbol string) (out *GetTokenListResp, err error) {
	params := make(map[string]string)
	params["module"] = "account"
	params["action"] = "get_token_list"
	params["account"] = account

	// 代币名称
	if symbol != "" {
		params["symbol"] = symbol
	}

	err = api.call(params, &out)
	return
}

// 查询账户的RAM/CPU/NET等资源信息
func (api *API) GetAccountResourceInfo(account string) (out *GetAccountResourceInfoResp, err error) {
	params := make(map[string]string)
	params["module"] = "account"
	params["action"] = "get_account_resource_info"
	params["account"] = account

	err = api.call(params, &out)
	return
}

// 查询交易详情信息
func (api *API) GetTransactionDetailInfo(trxId string)	(out *GetTransactionDetailInfoResp, err error) {
	params := make(map[string]string)
	params["module"] = "transaction"
	params["action"] = "get_transaction_detail_info"
	params["trx_id"] = trxId

	err = api.call(params, &out)
	return
}

type apiResult struct {
	Errno 	int    `json:"errno"`
	ErrMsg  string `json:"errmsg"`
	Data 	interface{}		`json:"data"`
}

var ErrNotFound = errors.New("resource not found")

func (api *API) call(params map[string]string, out interface{}) error {
	req, err := http.NewRequest("GET", api.ApiUrl, nil)
	if err != nil {
		return fmt.Errorf("NewRequest: %s", err)
	}

	query := req.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	query.Add("apikey", api.apiKey)
	req.URL.RawQuery = query.Encode()

	for k, v := range api.Header {
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header[k] = append(req.Header[k], v...)
	}

	if api.Debug {
		// Useful when debugging API calls
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(string(requestDump))
		fmt.Println("")
	}
	
	resp, err := api.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %s", req.URL.String(), err)
	}
	defer resp.Body.Close()

	var cnt bytes.Buffer
	_, err = io.Copy(&cnt, resp.Body)
	if err != nil {
		return fmt.Errorf("Copy: %s", err)
	}

	if api.Debug {
		fmt.Println("RESPONSE:")
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(cnt.String())
		fmt.Println("-------------------------------")
		fmt.Printf("%q\n", responseDump)
		fmt.Println("")
	}

	var result apiResult
	result.Data = out
	if err := json.Unmarshal(cnt.Bytes(), &result); err != nil {
		return fmt.Errorf("Unmarshal: %s", err)
	}
	
	if result.Errno != 0 {
		return errors.New(api.errorMap[result.Errno])
	}

	return nil
}