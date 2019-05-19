package etherscan

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
	"strings"
)

type API struct {
	HttpClient 			*http.Client
	Header     			http.Header
	ApiUrl    			string
	Debug      			bool

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
		apiKey: apiKey,
		Header:   make(http.Header),
	}

	return api
}

// 查询ETH交易信息列表
func (api *API) GetNormalTransactions(address ...string) (out []*GetNormalTransactionsResp, err error) {
	params := make(map[string]string)
	params["module"] = "account"
	params["action"] = "txlist"
	params["address"] = strings.Join(address, ",")
	
	err = api.call(params, &out)
	return
}

// 查询ERC20代币交易信息列表
func (api *API) GetERC20TokenTransactions(address, contractAddress string) (out []*GetERC20TokenTransactionsResp, err error) {
	params := make(map[string]string)
	params["module"] = "account"
	params["action"] = "tokentx"
	params["address"] = address
	params["contractaddress"] = contractAddress

	err = api.call(params, &out)
	return
}

// 查询交易信息状态
func (api *API) CheckTransactionReceiptStatus(tx string) (out *CheckTransactionReceiptStatusResp, err error) {
	params := make(map[string]string)
	params["module"] = "transaction"
	params["action"] = "gettxreceiptstatus"
	params["txhash"] = tx

	err = api.call(params, &out)
	return
}

type apiResult struct {
	Status 	string    `json:"status"`
	Message  string `json:"message"`
	Result 	interface{}		`json:"result"`
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
	result.Result = out
	if err := json.Unmarshal(cnt.Bytes(), &result); err != nil {
		return fmt.Errorf("Unmarshal: %s", err)
	}

	return nil
}