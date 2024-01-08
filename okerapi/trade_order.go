package okerapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

type TickerClosePositionReq struct {
	InstId  string `json:"instId"`
	UgnMode string `json:"mgnMode"`
	Ccy     string `json:"ccy"`
	PosSide string `json:"posSide"`
}

type TickerClosePositionItem struct {
	InstId  string `json:"instId"`
	PosSide string `json:"posSide"`
}

type TickerClosePositionRsp struct {
	Code string                    `json:"code"`
	Msg  string                    `json:"msg"`
	Data []TickerClosePositionItem `json:"data"`
}

// TickerClosePosition order_type 1是现货买单，2是期货买单
func TickerClosePosition(instid string, order_type string) (err error) {

	// body json 构造
	order_req := new(TickerClosePositionReq)
	order_req.InstId = instid
	order_req.UgnMode = "cross"

	if order_type == "MARGIN" {
		order_req.PosSide = "net"
		order_req.Ccy = "USDT"
	} else if order_type == "SWAP" {
		order_req.PosSide = "short"
	} else {
		fmt.Println("order_type undefined:", order_type)
		return err
	}

	reqBodyByte, err := json.Marshal(order_req)
	req, err := internal.InitUserEnvPost("/api/v5/trade/close-position", string(reqBodyByte))
	req.URL.Path = "/api/v5/trade/close-position"
	fmt.Println(string(reqBodyByte))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))
	var trade_rsp TickerClosePositionRsp
	err = json.Unmarshal(body, &trade_rsp)
	if err != nil {
		fmt.Println(err)
		return
	}

	if trade_rsp.Code == "0" {
		for _, order := range trade_rsp.Data {
			fmt.Println(order.InstId, order.PosSide, " order success")
		}
	}

	return err
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

	return ordId, internal.RETURN_SUCCESS
}

// TickersTradeOrder 交易接口
func TickersTradeOrder(
	instid string,
	instrument map[string]*InstrumentsDetail,
	last_spot_price string) (err error) {

	inst_swap_id := instid
	instid_array := strings.Split(inst_swap_id, "-")
	inst_spot_id := instid_array[0] + "-USDT"

	_, spot_ok := instrument[inst_spot_id]
	_, swap_ok := instrument[inst_swap_id]
	if spot_ok && swap_ok {
		spot_min_sz := instrument[inst_spot_id].MinSz
		swap_min_sz := instrument[inst_swap_id].MinSz
		fmt.Println("instrument find success", spot_min_sz, swap_min_sz)

		// 按照合约价格计算币币数量,合约价格默认60u,现货对应60u价格
		// 市价单，sz参数对应usdt的价格；限价单，sz参数对应币种价格
		if swap_min_sz == "1" {
			fmt.Println("instrument min sz == 1")
			// 同时提交现货和期货的买单
			TradeOrder(inst_spot_id, "1", "60")
			TradeOrder(inst_swap_id, "2", swap_min_sz)
		}
	} else {
		fmt.Println(inst_spot_id, spot_ok, inst_swap_id, swap_ok)
	}

	return err
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

	req, _ := http.NewRequest("GET", "https://www.okex.com", nil)
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

	return internal.RETURN_SUCCESS
}
