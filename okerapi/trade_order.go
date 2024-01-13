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

type TradeOrderRspItem struct {
	ClOrdId string `json:"clOrdId"`
	OrdId   string `json:"ordId"`
	Tag     string `json:"tag"`
	SCode   string `json:"sCode"`
	SMsg    string `json:"sMsg"`
}

type TradeOrderRsp struct {
	Code string              `json:"code"`
	Msg  string              `json:"msg"`
	Data []TradeOrderRspItem `json:"data"`
}

type TradeOrderReq struct {
	InstId  string `json:"instId"`
	TdMode  string `json:"tdMode"`
	Ccy     string `json:"ccy"`
	Side    string `json:"side"`
	PosSide string `json:"posSide"`
	OrdType string `json:"ordType"`
	Sz      string `json:"sz"`
	Px      string `json:"px"`
}

// TradeOrder order_type 1是现货买单，2是期货买单
func TradeOrder(instid string, orderType string, orderSz string) (string, string) {
	var ordId string

	// body json 构造
	tradeOrderReq := new(TradeOrderReq)
	tradeOrderReq.InstId = instid
	tradeOrderReq.TdMode = "cross"
	tradeOrderReq.OrdType = "market"
	tradeOrderReq.Sz = orderSz

	if orderType == "1" {
		tradeOrderReq.Side = "buy"
		tradeOrderReq.PosSide = "long"
		tradeOrderReq.Ccy = "USDT"
	} else if orderType == "2" {
		tradeOrderReq.Side = "sell"
		tradeOrderReq.PosSide = "short"
	} else {
		logInfo := fmt.Sprintf("TradeOrder orderType undefined: %s", orderType)
		internal.PrintDebugLogToFile(logInfo)
		return ordId, internal.TRADE_ORDER_REQ_ERROR
	}

	reqBodyByte, err := json.Marshal(tradeOrderReq)
	req, err := internal.InitUserEnvPost("/api/v5/trade/order", string(reqBodyByte))
	req.URL.Path = "/api/v5/trade/order"
	// fmt.Println(string(reqBodyByte))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(body))
	var tradeOrderRsp TradeOrderRsp
	err = json.Unmarshal(body, &tradeOrderRsp)
	if err != nil {
		logInfo := fmt.Sprintf("TradeOrder Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return ordId, internal.TRADE_ORDER_UNMARSHAL_RSP_ERROR
	}

	if tradeOrderRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("tradeOrderRsp err, code: %s, msg: %s", tradeOrderRsp.Code, tradeOrderRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return ordId, internal.TRADE_ORDER_POST_RSP_ERROR
	}

	if len(tradeOrderRsp.Data) != 1 {
		logInfo := fmt.Sprintf("tradeOrderRsp.Data counts exp: %d", len(tradeOrderRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return ordId, internal.TRADE_ORDER_POST_DATA_ERROR
	}

	if tradeOrderRsp.Data[0].SCode != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("tradeOrderRsp.Data[0].SCode != 0: %s", tradeOrderRsp.Data[0].SCode)
		internal.PrintDebugLogToFile(logInfo)
		return ordId, internal.TRADE_ORDER_POST_DATA_CODE_ERROR
	}

	ordId = tradeOrderRsp.Data[0].OrdId
	logInfo := fmt.Sprintf("trade order success: %s", ordId)
	internal.PrintDebugLogToFile(logInfo)

	logTradeInfo := fmt.Sprintf("%s|%s|BUY LONG SUCCESS", instid, ordId)
	internal.PrintTradeLogToFile(logTradeInfo)

	// 睡眠等待3秒成交，模拟盘10秒
	time.Sleep(10 * time.Second)

	return ordId, internal.RETURN_SUCCESS
}

type OrderDetailReq struct {
	InstId string `json:"instId"`
	OrdId  string `json:"ordId"`
}

type OrderDetail struct {
	InstId string `json:"instId"`
	State  string `json:"state"`
	AvgPx  string `json:"avgPx"`
}

type OrderDetailRsp struct {
	Code string        `json:"code"`
	Msg  string        `json:"msg"`
	Data []OrderDetail `json:"data"`
}

func GetOrderInfoOfAvgPx(instId string, ordId string) (string, string) {
	var avgPx string

	req, _ := http.NewRequest("GET", internal.ROOT_PATH, nil)
	req.URL.RawQuery = ""
	req.URL.Path = "/api/v5/trade/order"
	request_url := req.URL.Query()
	request_url.Add("instId", instId)
	request_url.Add("ordId", ordId)
	req.URL.RawQuery = request_url.Encode()
	// fmt.Println("order before url: " + req.URL.String())
	// fmt.Println("order RawQuery: " + req.URL.RawQuery)

	pathUrl := "/api/v5/trade/order?" + req.URL.RawQuery
	req = internal.InitUserEnv(pathUrl)
	req.URL.Path = "/api/v5/trade/order"
	req.URL.RawQuery = request_url.Encode()

	// fmt.Println("order after url: " + req.URL.String())
	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(body))
	var orderDetailRsp OrderDetailRsp
	err := json.Unmarshal(body, &orderDetailRsp)
	if err != nil {
		logInfo := fmt.Sprintf("getOrderInfoOfAvgPx Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return avgPx, internal.GET_ORDER_INFO_UNMARSHAL_RSP_ERROR
	}

	if orderDetailRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("getOrderInfoOfAvgPx err, code: %s, msg: %s", orderDetailRsp.Code, orderDetailRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return avgPx, internal.GET_ORDER_INFO_POST_RSP_ERROR
	}

	if len(orderDetailRsp.Data) != 1 {
		logInfo := fmt.Sprintf("orderDetailRsp.Data counts exp: %d", len(orderDetailRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return avgPx, internal.GET_ORDER_INFO_POST_DATA_ERROR
	}

	// 未成交 或 未完全成交 的状态还未处理, 先打印该情况，逻辑上市价单应该是秒成交
	if orderDetailRsp.Data[0].State != "filled" {
		logInfo := fmt.Sprintf("orderDetailRsp.Data[0].State != filled: %s", orderDetailRsp.Data[0].State)
		internal.PrintDebugLogToFile(logInfo)

		logTradeInfo := fmt.Sprintf("%s|%s|BUY FAIL", instId, orderDetailRsp.Data[0].State)
		internal.PrintTradeLogToFile(logTradeInfo)
		return avgPx, internal.GET_ORDER_INFO_POST_DATA_STATE_ERROR
	}

	avgPx = orderDetailRsp.Data[0].AvgPx
	logInfo := fmt.Sprintf("all order finish: %s", avgPx)
	internal.PrintDebugLogToFile(logInfo)

	return avgPx, internal.RETURN_SUCCESS
}

type OrderAlgoReq struct {
	InstId        string `json:"instId"`
	TdMode        string `json:"tdMode"`
	Ccy           string `json:"ccy"`
	Side          string `json:"side"`
	PosSide       string `json:"posSide"`
	OrdType       string `json:"ordType"`
	Sz            string `json:"sz"`
	SlTriggerPx   string `json:"slTriggerPx"`
	SlOrdPx       string `json:"slOrdPx"`
	CloseFraction string `json:"closeFraction"`
}

type OrderAlgo struct {
	AlgoId string `json:"algoId"`
	SCode  string `json:"sCode"`
	SMsg   string `json:"sMsg"`
}

type OrderAlgoRsp struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data []OrderAlgo `json:"data"`
}

// TradeSLOrder 止损单
func TradeSLOrder(instid string, orderPrice string, orderCounts string) string {
	// body json 构造
	orderAlgoReq := new(OrderAlgoReq)
	orderAlgoReq.InstId = instid
	orderAlgoReq.TdMode = "cross"
	orderAlgoReq.OrdType = "conditional"
	// orderAlgoReq.Sz = orderCounts
	orderAlgoReq.Side = "sell"
	orderAlgoReq.PosSide = "long"
	orderAlgoReq.Ccy = "USDT"
	orderAlgoReq.SlTriggerPx = orderPrice
	orderAlgoReq.SlOrdPx = "-1"
	orderAlgoReq.CloseFraction = "1"

	reqBodyByte, err := json.Marshal(orderAlgoReq)
	req, err := internal.InitUserEnvPost("/api/v5/trade/order-algo", string(reqBodyByte))
	req.URL.Path = "/api/v5/trade/order-algo"
	// fmt.Println(string(reqBodyByte))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(body))
	var orderAlgoRsp OrderAlgoRsp
	err = json.Unmarshal(body, &orderAlgoRsp)
	if err != nil {
		logInfo := fmt.Sprintf("TradeSLOrder Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return internal.TRADE_SL_ORDER_UNMARSHAL_RSP_ERROR
	}

	if orderAlgoRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("orderAlgoRsp err, code: %s, msg: %s", orderAlgoRsp.Code, orderAlgoRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return internal.TRADE_SL_ORDER_POST_RSP_ERROR
	}

	if len(orderAlgoRsp.Data) != 1 {
		logInfo := fmt.Sprintf("orderAlgoRsp.Data counts exp: %d", len(orderAlgoRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return internal.TRADE_SL_ORDER_POST_DATA_ERROR
	}

	logInfo := fmt.Sprintf("trade sl order success: %s", orderAlgoRsp.Data[0].AlgoId)
	internal.PrintDebugLogToFile(logInfo)

	logTradeInfo := fmt.Sprintf("%s|%s|SELL LONG SUCCESS", instid, orderAlgoRsp.Data[0].AlgoId)
	internal.PrintTradeLogToFile(logTradeInfo)

	return internal.RETURN_SUCCESS
}

// 已经开仓过的产品，平仓前不再开仓，以及4H周期内也不再开仓
func JudgeTradeCondition(instId string) bool {
	// 拉取持仓信息
	accountPositionList, ret := GetAccountPosition()
	if ret == internal.GET_ACCOUNT_POSITION_UNMARSHAL_RSP_ERROR || ret == internal.GET_ACCOUNT_POSITION_POST_RSP_ERROR {
		logInfo := fmt.Sprintf("GetAccountPosition err: %s", ret)
		internal.PrintDebugLogToFile(logInfo)
		return false
	}

	// 持仓超过20个，不再开仓
	if len(accountPositionList) >= 3 {
		logInfo := fmt.Sprintf("JudgeTradeCondition false:len(accountPositionList) >= 3")
		internal.PrintDebugLogToFile(logInfo)
		return false
	}

	// 在开仓列表不再重复开仓
	for _, eachCase := range accountPositionList {
		if eachCase.InstId == instId {
			logInfo := fmt.Sprintf("JudgeTradeCondition false:eachCase.InstId == instId")
			internal.PrintDebugLogToFile(logInfo)
			return false
		}
	}

	// 拉取历史持仓信息
	accountPositionHistoryList, ret := GetAccountPositionHistory()
	if ret == internal.GET_ACCOUNT_POSITION_HISTORY_UNMARSHAL_RSP_ERROR || ret == internal.GET_ACCOUNT_POSITION_HISTORY_POST_RSP_ERROR {
		logInfo := fmt.Sprintf("GetAccountPositionHistory err: %s", ret)
		internal.PrintDebugLogToFile(logInfo)
		return false
	}

	// 拉取间隔1次/10
	logInfo := fmt.Sprintf("GetAccountPositionHistory wait time.Sleep(10 * time.Second): %d", len(accountPositionHistoryList))
	internal.PrintDebugLogToFile(logInfo)
	time.Sleep(10 * time.Second)

	// 4H周期内也不再开仓
	for _, eachCase := range accountPositionHistoryList {
		if eachCase.InstId == instId {
			logInfo := fmt.Sprintf("JudgeTradeCondition 4H false:eachCase.InstId == instId")
			internal.PrintDebugLogToFile(logInfo)
			return false
		}
	}

	return true
}

type AccountPositionHistory struct {
	InstId string `json:"instId"`
	CTime  string `json:"cTime"`
	UTime  string `json:"uTime"`
}

type AccountPositionHistoryRsp struct {
	Code string                   `json:"code"`
	Msg  string                   `json:"msg"`
	Data []AccountPositionHistory `json:"data"`
}

func GetAccountPositionHistory() ([]AccountPositionHistory, string) {
	var accountPositionHistoryList []AccountPositionHistory
	req, _ := http.NewRequest("GET", internal.ROOT_PATH, nil)
	accountPostionPath := "/api/v5/account/positions-history"

	// 四个小时前
	tNow := time.Now()
	tNowBefore4H := tNow.Add(-1 * time.Hour)
	before4H := strconv.FormatInt(tNowBefore4H.UnixNano()/int64(time.Millisecond), 10)

	req.URL.Path = accountPostionPath
	request_url := req.URL.Query()
	request_url.Add("instType", "SWAP")
	request_url.Add("before", before4H)
	// request_url.Add("limit", "1")
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
	var accountPositionHistoryRsp AccountPositionHistoryRsp
	err := json.Unmarshal(body, &accountPositionHistoryRsp)
	if err != nil {
		logInfo := fmt.Sprintf("CheckAccountPositionsHistory Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return accountPositionHistoryList, internal.GET_ACCOUNT_POSITION_HISTORY_UNMARSHAL_RSP_ERROR
	}

	if accountPositionHistoryRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("CheckAccountPositionsHistory err, code: %s, msg: %s", accountPositionHistoryRsp.Code, accountPositionHistoryRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return accountPositionHistoryList, internal.GET_ACCOUNT_POSITION_HISTORY_POST_RSP_ERROR
	}

	// 开仓数量逻辑上和止损策略单数量一样，不过考虑非事务一致关系，只对不一致的产品做日志输出
	if len(accountPositionHistoryRsp.Data) < 1 {
		logInfo := fmt.Sprintf("len(accountPositionHistoryRsp.Data) < 1: %d", len(accountPositionHistoryRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return accountPositionHistoryList, internal.GET_ACCOUNT_POSITION_HISTORY_POST_DATA_ERROR
	}

	accountPositionHistoryList = accountPositionHistoryRsp.Data

	// unixTime, _ := strconv.ParseInt(accountPositionHistoryList[len(accountPositionHistoryRsp.Data)-1].CTime, 10, 64)
	// unixTimeTmType := time.Unix(0, unixTime*int64(time.Millisecond))
	// cTimeformat := unixTimeTmType.Format("2006-01-02 15:04:05")

	// unixTime, _ = strconv.ParseInt(accountPositionHistoryList[len(accountPositionHistoryRsp.Data)-1].UTime, 10, 64)
	// unixTimeTmType = time.Unix(0, unixTime*int64(time.Millisecond))
	// uTimeformat := unixTimeTmType.Format("2006-01-02 15:04:05")

	// tempInstId := accountPositionHistoryList[len(accountPositionHistoryRsp.Data)-1].InstId
	// logInfo := fmt.Sprintf("len(accountPositionHistoryRsp.Data):%d, %s, %s, %s", len(accountPositionHistoryRsp.Data), tempInstId, cTimeformat, uTimeformat)
	// internal.PrintDebugLogToFile(logInfo)

	return accountPositionHistoryList, internal.RETURN_SUCCESS
}
