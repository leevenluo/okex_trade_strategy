package okerapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"trade_strategy/internal"
)

type MarketTicker struct {
	InstId string `json:"instId"`
	Last   string `json:"last"`
}

type MarketTickerRsp struct {
	Code string         `json:"code"`
	Msg  string         `json:"msg"`
	Data []MarketTicker `json:"data"`
}

// 获取指定产品行情信息
func GetMarketTicker(instId string) (MarketTicker, string) {
	req, _ := http.NewRequest("GET", internal.ROOT_PATH, nil)
	accountPostionPath := "/api/v5/market/ticker"

	req.URL.Path = accountPostionPath
	request_url := req.URL.Query()
	request_url.Add("instId", instId)
	req.URL.RawQuery = request_url.Encode()

	pathUrl := accountPostionPath + "?" + req.URL.RawQuery
	req = internal.InitUserEnv(pathUrl)
	req.URL.Path = accountPostionPath
	req.URL.RawQuery = request_url.Encode()

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var marketTickerRsp MarketTickerRsp
	var marketTicker MarketTicker
	err := json.Unmarshal(body, &marketTickerRsp)
	if err != nil {
		logInfo := fmt.Sprintf("GetMarketTicker Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return marketTicker, internal.GET_MARKET_TICKER_UNMARSHAL_RSP_ERROR
	}

	if marketTickerRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("GetMarketTicker err, code: %s, msg: %s", marketTickerRsp.Code, marketTickerRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return marketTicker, internal.GET_MARKET_TICKER_POST_RSP_ERROR
	}

	if len(marketTickerRsp.Data) != 1 {
		logInfo := fmt.Sprintf("marketTickerRsp.Data counts exp: %d", len(marketTickerRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return marketTicker, internal.GET_MARKET_TICKER_POST_DATA_ERROR
	}

	return marketTickerRsp.Data[0], internal.RETURN_SUCCESS
}
