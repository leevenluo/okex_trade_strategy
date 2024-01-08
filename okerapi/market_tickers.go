package okerapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"trade_strategy/internal"
)

// ProductDetail 资产列表struct定义
type ProductDetail struct {
	InstType  string `json:"instType"`
	InstId    string `json:"instId"`
	Last      string `json:"last"`
	LastSz    string `json:"lastSz"`
	AskPx     string `json:"askPx"`
	AskSz     string `json:"askSz"`
	BidPx     string `json:"bidPx"`
	BidSz     string `json:"bidSz"`
	Open24h   string `json:"open24h"`
	High24h   string `json:"high24h"`
	Low24h    string `json:"low24h"`
	VolCcy24h string `json:"volCcy24h"`
	Vol24h    string `json:"vol24h"`
	SodUtc0   string `json:"sodUtc0"`
	SodUtc8   string `json:"sodUtc8"`
	Ts        string `json:"ts"`
}

type ProductInfo struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data []ProductDetail `json:"data"`
}

type ProductList struct {
	InstId string
}

// GetMarketTickers 获取产品列表
func GetMarketTickers(req *http.Request) ([]ProductList, string) {
	var productList []ProductList
	req.URL.RawQuery = ""
	req.URL.Path = "/api/v5/market/tickers"
	request_url := req.URL.Query()
	request_url.Add("instType", "SWAP")
	req.URL.RawQuery = request_url.Encode()

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))

	var productInfo ProductInfo
	err := json.Unmarshal(body, &productInfo)
	if err != nil {
		logInfo := fmt.Sprintf("GetMarketTickers Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return productList, internal.GET_MARKET_TICKERS_UNMARSHAL_RSP_ERROR
	}

	if productInfo.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("orderDetailRsp err, code: %s, msg: %s", productInfo.Code, productInfo.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return productList, internal.GET_ORDER_INFO_POST_RSP_ERROR
	}

	unvalidSwapTicker := []string{}
	for _, each_item := range productInfo.Data {
		// 只过滤u本位合约，规则BTC-USDT-SWAP，区别BTC-USD-SWAP的币本位
		instid_array := strings.Split(each_item.InstId, "-")
		// if len(instid_array) >= 3 && instid_array[0] == "BTC" && instid_array[1] == "USDT" {
		if len(instid_array) >= 3 && instid_array[1] == "USDT" {
			var productitem ProductList
			productitem.InstId = each_item.InstId
			productList = append(productList, productitem)
		} else {
			unvalidSwapTicker = append(unvalidSwapTicker, each_item.InstId)
		}
	}

	logInfo := fmt.Sprintf("len productList :", len(productList))
	internal.PrintDebugLogToFile(logInfo)
	//fmt.Println(unvalidSwapTicker)

	return productList, internal.RETURN_SUCCESS
}
