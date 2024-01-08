package okerapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"trade_strategy/internal"
)

type CandlesRspData struct {
	Ts          string `json:"ts"`
	OpenPrice   string `json:"o"`
	HighPrice   string `json:"h"`
	LowPrice    string `json:"l"`
	ClosePrice  string `json:"c"`
	Vol         string `json:"vol"`
	VolCcy      string `json:"volCcy"`
	VolCcyQuote string `json:"volCcyQuote"`
	Confirm     string `json:"confirm"`
	FormatTime  string `json:"formatTime"`
}

type GetCandlesWrapRsp struct {
	Code string           `json:"code"`
	Msg  string           `json:"msg"`
	Data []CandlesRspData `json:"data"`
}

type GetCandlesRsp struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

type GetCandlesResult struct {
	InstId               string
	Get15mCandlesWrapRsp GetCandlesWrapRsp
	Get4HCandlesWrapRsp  GetCandlesWrapRsp
}

func GetCandles(instId string) (*GetCandlesResult, string) {
	getCandlesResult := new(GetCandlesResult)
	// 获取最近15m和4H的K线
	get15mCandlesWrapRsp, ret15m := getPeriodCandles("15m", instId)
	get4HCandlesWrapRsp, ret4H := getPeriodCandles("4H", instId)
	if ret15m != internal.RETURN_SUCCESS || ret4H != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("getPeriodCandles %s err: %s, %s", instId, ret15m, ret4H)
		fmt.Println(logInfo)
		return getCandlesResult, internal.GET_CANDLES_RESULT_ERROR
	}

	getCandlesResult.InstId = instId
	getCandlesResult.Get15mCandlesWrapRsp = get15mCandlesWrapRsp
	getCandlesResult.Get4HCandlesWrapRsp = get4HCandlesWrapRsp

	return getCandlesResult, internal.RETURN_SUCCESS
}

func getPeriodCandles(period string, instId string) (GetCandlesWrapRsp, string) {
	var getCandlesWrapRsp GetCandlesWrapRsp

	// req.URL.RawQuery = ""
	// req.URL.Path = "/api/v5/market/candles"
	// request_url := req.URL.Query()
	// request_url.Add("instId", instId)
	// request_url.Add("bar", period)
	// request_url.Add("limit", "1")
	// req.URL.RawQuery = request_url.Encode()

	req, _ := http.NewRequest("GET", internal.ROOT_PATH, nil)
	accountPostionPath := "/api/v5/market/candles"

	req.URL.Path = accountPostionPath
	request_url := req.URL.Query()
	request_url.Add("instId", instId)
	request_url.Add("bar", period)
	request_url.Add("limit", "1")
	req.URL.RawQuery = request_url.Encode()

	pathUrl := accountPostionPath + "?" + req.URL.RawQuery
	req = internal.InitUserEnv(pathUrl)
	req.URL.Path = accountPostionPath
	req.URL.RawQuery = request_url.Encode()

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))

	var getCandlesRsp GetCandlesRsp
	err := json.Unmarshal(body, &getCandlesRsp)
	if err != nil {
		logInfo := fmt.Sprintf("getPeriodCandles Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return getCandlesWrapRsp, internal.GET_PERIOD_CANDLES_UNMARSHAL_RSP_ERROR
	}

	if getCandlesRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("getPeriodCandles err, code: %s, msg: %s", getCandlesRsp.Code, getCandlesRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return getCandlesWrapRsp, internal.GET_PERIOD_CANDLES_POST_RSP_ERROR
	}

	if len(getCandlesRsp.Data) != 1 {
		logInfo := fmt.Sprintf("getPeriodCandles.Data counts exp: %d", len(getCandlesRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return getCandlesWrapRsp, internal.GET_PERIOD_CANDLES_POST_DATA_ERROR
	}

	getCandlesWrapRsp.Code = getCandlesRsp.Code
	getCandlesWrapRsp.Msg = getCandlesRsp.Msg
	for _, eachCandlesData := range getCandlesRsp.Data {
		var candlesRspData CandlesRspData
		candlesRspData.Ts = eachCandlesData[0]
		candlesRspData.OpenPrice = eachCandlesData[1]
		candlesRspData.HighPrice = eachCandlesData[2]
		candlesRspData.LowPrice = eachCandlesData[3]
		candlesRspData.ClosePrice = eachCandlesData[4]

		unixTime, _ := strconv.ParseInt(eachCandlesData[0], 10, 64)
		unixTimeTmType := time.Unix(0, unixTime*int64(time.Millisecond))
		candlesRspData.FormatTime = unixTimeTmType.Format("2006-01-02 15:04:05")

		getCandlesWrapRsp.Data = append(getCandlesWrapRsp.Data, candlesRspData)
	}

	return getCandlesWrapRsp, internal.RETURN_SUCCESS
}

// 获取最近N条K线
func getPeriodCandlesBatch(period string, limitCounts string, instId string) (GetCandlesWrapRsp, string) {
	var getCandlesWrapRsp GetCandlesWrapRsp

	// req.URL.RawQuery = ""
	// req.URL.Path = "/api/v5/market/candles"
	// request_url := req.URL.Query()
	// request_url.Add("instId", instId)
	// request_url.Add("bar", period)
	// request_url.Add("limit", "1")
	// req.URL.RawQuery = request_url.Encode()

	req, _ := http.NewRequest("GET", internal.ROOT_PATH, nil)
	accountPostionPath := "/api/v5/market/candles"

	req.URL.Path = accountPostionPath
	request_url := req.URL.Query()
	request_url.Add("instId", instId)
	request_url.Add("bar", period)
	request_url.Add("limit", limitCounts)
	req.URL.RawQuery = request_url.Encode()

	pathUrl := accountPostionPath + "?" + req.URL.RawQuery
	req = internal.InitUserEnv(pathUrl)
	req.URL.Path = accountPostionPath
	req.URL.RawQuery = request_url.Encode()

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))

	var getCandlesRsp GetCandlesRsp
	err := json.Unmarshal(body, &getCandlesRsp)
	if err != nil {
		logInfo := fmt.Sprintf("getPeriodCandles Unmarshal err: %s", err.Error())
		fmt.Println(logInfo)
		return getCandlesWrapRsp, internal.GET_PERIOD_CANDLES_UNMARSHAL_RSP_ERROR
	}

	if getCandlesRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("getPeriodCandles err, code: %s, msg: %s", getCandlesRsp.Code, getCandlesRsp.Msg)
		fmt.Println(logInfo)
		return getCandlesWrapRsp, internal.GET_PERIOD_CANDLES_POST_RSP_ERROR
	}

	if len(getCandlesRsp.Data) < 1 {
		logInfo := fmt.Sprintf("getPeriodCandles.Data counts exp: %d", len(getCandlesRsp.Data))
		fmt.Println(logInfo)
		return getCandlesWrapRsp, internal.GET_PERIOD_CANDLES_POST_DATA_ERROR
	}

	getCandlesWrapRsp.Code = getCandlesRsp.Code
	getCandlesWrapRsp.Msg = getCandlesRsp.Msg
	for _, eachCandlesData := range getCandlesRsp.Data {
		var candlesRspData CandlesRspData
		candlesRspData.Ts = eachCandlesData[0]
		candlesRspData.OpenPrice = eachCandlesData[1]
		candlesRspData.HighPrice = eachCandlesData[2]
		candlesRspData.LowPrice = eachCandlesData[3]
		candlesRspData.ClosePrice = eachCandlesData[4]

		unixTime, _ := strconv.ParseInt(eachCandlesData[0], 10, 64)
		unixTimeTmType := time.Unix(0, unixTime*int64(time.Millisecond))
		candlesRspData.FormatTime = unixTimeTmType.Format("2006-01-02 15:04:05")

		getCandlesWrapRsp.Data = append(getCandlesWrapRsp.Data, candlesRspData)
	}

	logInfo := fmt.Sprintf("getPeriodCandles.Data counts: %d, new counts: %d", len(getCandlesRsp.Data), len(getCandlesWrapRsp.Data))
	fmt.Println(logInfo)
	return getCandlesWrapRsp, internal.RETURN_SUCCESS
}
