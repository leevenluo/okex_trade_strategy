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

// 针对已经建仓产品，定时轮询
func CheckAccountPositions() {
	// 初始化启动时间戳，为当天0点unix时间戳
	// var last_time int64
	// last_time, ret := internal.InitZeroTimestamp()
	// if ret != internal.RETURN_SUCCESS {
	// 	logInfo := fmt.Sprintf("internal.InitZeroTimestamp err: %s", ret)
	// 	internal.PrintDebugLogToFile(logInfo)
	// 	return
	// }

	// firstRun := true
	// var interval int64 = 30

	for true {
		// 拉取持仓信息
		accountPositionList, ret := GetAccountPosition()
		if ret != internal.RETURN_SUCCESS {
			logInfo := fmt.Sprintf("GetAccountPosition err: %s", ret)
			internal.PrintDebugLogToFile(logInfo)
			continue
		}

		// 遍历持仓产品
		for _, eachCase := range accountPositionList {
			time.Sleep(1 * time.Second)
			// 产品没有止损单，打印异常信息后退出
			if len(eachCase.CloseOrderAlgoList) < 1 {
				logInfo := fmt.Sprintf("No CloseOrderAlgoList err: %s", eachCase.InstId)
				internal.PrintDebugLogToFile(logInfo)
				continue
			}

			// 有止损单，判断当前行情是否需要调整止损单价格
			status, newSLPrice, ret := CheckAdjustSLPrice(eachCase)
			if ret != internal.RETURN_SUCCESS {
				logInfo := fmt.Sprintf("CheckAdjustSLPrice err: %s", ret)
				internal.PrintDebugLogToFile(logInfo)
				continue
			}

			logInfo := fmt.Sprintf("CheckAdjustSLPrice status: %t", status)
			internal.PrintDebugLogToFile(logInfo)

			if status {
				// 调整止损单价格
				ret = AdjustSLPrice(eachCase, newSLPrice)
				if ret != internal.RETURN_SUCCESS {
					logInfo := fmt.Sprintf("AdjustSLPrice err: %s", ret)
					internal.PrintDebugLogToFile(logInfo)
					continue
				}
			}

			// 是否达到出场条件，达到后出场
			status, ret = CheckCloseCondition(eachCase)
			if ret != internal.RETURN_SUCCESS {
				logInfo := fmt.Sprintf("CheckCloseCondition err: %s", ret)
				internal.PrintDebugLogToFile(logInfo)
				continue
			}

			if status {
				// 出场平仓
				ret = ClosePosition(eachCase)
				if ret != internal.RETURN_SUCCESS {
					logInfo := fmt.Sprintf("ClosePosition err: %s", ret)
					internal.PrintDebugLogToFile(logInfo)
					continue
				}

				// 计算盈亏金额
				ret = CalcTradeProfit(eachCase.InstId)
				if ret != internal.RETURN_SUCCESS {
					logInfo := fmt.Sprintf("CalcTradeProfit err: %s", ret)
					internal.PrintDebugLogToFile(logInfo)
					continue
				}
			}
		}

		// 周期判断轮询
		// for true {
		// 	var status bool
		// 	status, err = internal.TimeIntervalSuccess(&last_time, interval, 2)
		// 	if firstRun {
		// 		time.Sleep(60 * time.Second)
		// 		firstRun = false
		// 	} else {
		// 		if status {
		// 			break
		// 		} else {
		// 			logInfo := fmt.Sprintf("market ticker period interval check, interval(s) =", interval)
		// 			internal.PrintDebugLogToFile(logInfo)
		// 		}
		// 		time.Sleep(100 * time.Second)
		// 	}
		// }
	}
}

type CloseOrderAlgo struct {
	AlgoId string `json:"algoId"`
}

type AccountPosition struct {
	InstId             string           `json:"instId"`
	AvgPx              string           `json:"avgPx"`
	CloseOrderAlgoList []CloseOrderAlgo `json:"closeOrderAlgo"`
}

type AccountPositionRsp struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Data []AccountPosition `json:"data"`
}

func GetAccountPosition() ([]AccountPosition, string) {
	var accountPositionList []AccountPosition
	req, _ := http.NewRequest("GET", internal.ROOT_PATH, nil)
	accountPostionPath := "/api/v5/account/positions"

	req.URL.Path = accountPostionPath
	request_url := req.URL.Query()
	request_url.Add("instType", "SWAP")
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
	var accountPositionRsp AccountPositionRsp
	err := json.Unmarshal(body, &accountPositionRsp)
	if err != nil {
		logInfo := fmt.Sprintf("CheckAccountPositions Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return accountPositionList, internal.GET_ACCOUNT_POSITION_UNMARSHAL_RSP_ERROR
	}

	if accountPositionRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("CheckAccountPositions err, code: %s, msg: %s", accountPositionRsp.Code, accountPositionRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return accountPositionList, internal.GET_ACCOUNT_POSITION_POST_RSP_ERROR
	}

	// 开仓数量逻辑上和止损策略单数量一样，不过考虑非事务一致关系，只对不一致的产品做日志输出
	if len(accountPositionRsp.Data) < 1 {
		logInfo := fmt.Sprintf("len(accountPositionRsp.Data) < 1: %d", len(accountPositionRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return accountPositionList, internal.GET_ACCOUNT_POSITION_POST_DATA_ERROR
	}

	accountPositionList = accountPositionRsp.Data
	return accountPositionList, internal.RETURN_SUCCESS
}

func CheckAdjustSLPrice(accountPosition AccountPosition) (bool, string, string) {
	var newSLPrice string

	// 获取最新成交价
	marketTicker, ret := GetMarketTicker(accountPosition.InstId)
	if ret != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("CheckAdjustSLPrice GetMarketTicker err: %s", ret)
		internal.PrintDebugLogToFile(logInfo)
		return false, newSLPrice, internal.CHECK_ADJUST_SLPRICE_KLINE_ERROR
	}

	// 获取4H K线
	get4HCandlesWrapRsp, ret4H := getPeriodCandles("4H", accountPosition.InstId)

	if ret4H != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("CheckAdjustSLPrice err, ret4H: %s", ret4H)
		internal.PrintDebugLogToFile(logInfo)
		return false, newSLPrice, internal.CHECK_ADJUST_SLPRICE_KLINE_ERROR
	}

	// 判断止损价格是否调整
	status, newSLPrice := JudgeSLPrice(marketTicker, accountPosition.AvgPx, get4HCandlesWrapRsp)

	return status, newSLPrice, internal.RETURN_SUCCESS
}

type AmendAlgoReq struct {
	InstId         string `json:"instId"`
	AlgoId         string `json:"algoId"`
	NewSlTriggerPx string `json:"newSlTriggerPx"`
}

type AmendAlgo struct {
	AlgoId string `json:"algoId"`
	SCode  string `json:"sCode"`
	SMsg   string `json:"sMsg"`
}

type AmendAlgoRsp struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data []AmendAlgo `json:"data"`
}

// 调整止损单价格
func AdjustSLPrice(accountPosition AccountPosition, newSLPrice string) string {
	var amendAlgoReq AmendAlgoReq
	amendAlgoReq.InstId = accountPosition.InstId
	amendAlgoReq.AlgoId = accountPosition.CloseOrderAlgoList[0].AlgoId
	amendAlgoReq.NewSlTriggerPx = newSLPrice

	reqBodyByte, err := json.Marshal(amendAlgoReq)
	req, err := internal.InitUserEnvPost("/api/v5/trade/amend-algos", string(reqBodyByte))
	req.URL.Path = "/api/v5/trade/amend-algos"
	fmt.Println(string(reqBodyByte))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))
	var amendAlgoRsp AmendAlgoRsp
	err = json.Unmarshal(body, &amendAlgoRsp)
	if err != nil {
		logInfo := fmt.Sprintf("AdjustSLPrice Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return internal.ADJUST_SL_RPICE_UNMARSHAL_RSP_ERROR
	}

	if amendAlgoRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("AdjustSLPrice err, code: %s, msg: %s", amendAlgoRsp.Code, amendAlgoRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return internal.ADJUST_SL_RPICE_POST_RSP_ERROR
	}

	if len(amendAlgoRsp.Data) != 1 {
		logInfo := fmt.Sprintf("AdjustSLPrice.Data counts exp: %d", len(amendAlgoRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return internal.ADJUST_SL_RPICE_POST_DATA_ERROR
	}

	if amendAlgoRsp.Data[0].SCode != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("AdjustSLPrice amend err: %s", amendAlgoRsp.Data[0].SCode)
		internal.PrintDebugLogToFile(logInfo)
		return internal.ADJUST_SL_RPICE_RESULT_ERROR
	}

	logInfo := fmt.Sprintf("AdjustSLPrice success: %s", amendAlgoRsp.Data[0].AlgoId)
	internal.PrintDebugLogToFile(logInfo)

	logTradeInfo := fmt.Sprintf("%s|%s|ADJUST SELL LONG SUCCESS", amendAlgoReq.InstId, amendAlgoRsp.Data[0].AlgoId)
	internal.PrintTradeLogToFile(logTradeInfo)

	return internal.RETURN_SUCCESS
}

// 是否达到出场条件，达到后出场
func CheckCloseCondition(accountPosition AccountPosition) (bool, string) {
	// 获取近20个15m K线
	get15mCandlesWrapBatchRsp, ret15m := getPeriodCandlesBatch("15m", "20", accountPosition.InstId)
	if ret15m != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("CheckCloseCondition getPeriodCandlesBatch err, ret15m: %s", ret15m)
		internal.PrintDebugLogToFile(logInfo)
		return false, internal.CHECK_ADJUST_SLPRICE_KLINE_ERROR
	}

	// 判断出场平仓条件是否满足
	status := JudgeClosePosition(get15mCandlesWrapBatchRsp)

	return status, internal.RETURN_SUCCESS
}

type ClosePositionReq struct {
	InstId  string `json:"instId"`
	PosSide string `json:"posSide"`
	MgnMode string `json:"mgnMode"`
	Ccy     string `json:"ccy"`
}

type ClosePositionInfo struct {
	InstId  string `json:"instId"`
	PosSide string `json:"posSide"`
}

type ClosePositionRsp struct {
	Code string              `json:"code"`
	Msg  string              `json:"msg"`
	Data []ClosePositionInfo `json:"data"`
}

// 出场平仓
func ClosePosition(accountPosition AccountPosition) string {
	var closePositionReq ClosePositionReq
	closePositionReq.InstId = accountPosition.InstId
	closePositionReq.PosSide = "long"
	closePositionReq.MgnMode = "cross"
	closePositionReq.Ccy = "USDT"

	reqBodyByte, err := json.Marshal(closePositionReq)
	req, err := internal.InitUserEnvPost("/api/v5/trade/close-position", string(reqBodyByte))
	req.URL.Path = "/api/v5/trade/close-position"
	// fmt.Println(string(reqBodyByte))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(body))
	var closePositionRsp ClosePositionRsp
	err = json.Unmarshal(body, &closePositionRsp)
	if err != nil {
		logInfo := fmt.Sprintf("ClosePosition Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return internal.CLOSE_POSITION_UNMARSHAL_RSP_ERROR
	}

	if closePositionRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("ClosePosition err, code: %s, msg: %s", closePositionRsp.Code, closePositionRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return internal.CLOSE_POSITION_POST_RSP_ERROR
	}

	if len(closePositionRsp.Data) != 1 {
		logInfo := fmt.Sprintf("ClosePosition.Data counts exp: %d", len(closePositionRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return internal.CLOSE_POSITION_POST_DATA_ERROR
	}

	logInfo := fmt.Sprintf("ClosePosition success: %s", closePositionRsp.Data[0].InstId)
	internal.PrintDebugLogToFile(logInfo)

	logTradeInfo := fmt.Sprintf("%s|%s|CLOSE POSITION SUCCESS", closePositionReq.InstId, closePositionRsp.Data[0].PosSide)
	internal.PrintTradeLogToFile(logTradeInfo)

	return internal.RETURN_SUCCESS
}

type TradeFills struct {
	InstId   string `json:"instId"`
	Side     string `json:"side"`
	Fee      string `json:"fee"`
	FillPnl  string `json:"fillPnl"`
	FillTime string `json:"fillTime"`
}

type TradeFillsRsp struct {
	Code string       `json:"code"`
	Msg  string       `json:"msg"`
	Data []TradeFills `json:"data"`
}

// 计算盈亏金额
func CalcTradeProfit(instId string) string {
	// 先睡眠10s再算平仓收益，正式环境换成3s
	time.Sleep(10 * time.Second)
	req, _ := http.NewRequest("GET", internal.ROOT_PATH, nil)
	accountPostionPath := "/api/v5/trade/fills"

	req.URL.Path = accountPostionPath
	request_url := req.URL.Query()
	request_url.Add("instType", "SWAP")
	request_url.Add("instId", instId)
	request_url.Add("limit", "2")
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
	var tradeFillsRsp TradeFillsRsp
	err := json.Unmarshal(body, &tradeFillsRsp)
	if err != nil {
		logInfo := fmt.Sprintf("CalcTradeProfit Unmarshal err: %s", err.Error())
		internal.PrintDebugLogToFile(logInfo)
		return internal.CALC_TRADE_PROFIT_UNMARSHAL_RSP_ERROR
	}

	if tradeFillsRsp.Code != internal.RETURN_SUCCESS {
		logInfo := fmt.Sprintf("CalcTradeProfit err, code: %s, msg: %s", tradeFillsRsp.Code, tradeFillsRsp.Msg)
		internal.PrintDebugLogToFile(logInfo)
		return internal.CALC_TRADE_PROFIT_POST_RSP_ERROR
	}

	if len(tradeFillsRsp.Data) != 2 {
		logInfo := fmt.Sprintf("tradeFillsRsp.Data counts exp: %d", len(tradeFillsRsp.Data))
		internal.PrintDebugLogToFile(logInfo)
		return internal.CALC_TRADE_PROFIT_POST_DATA_ERROR
	}

	// 通过最近sell平仓，次新bug建仓，且sell的成交时间小于1分钟，认为平仓收益计算符合逻辑
	unixTimestampMillis, _ := strconv.ParseInt(tradeFillsRsp.Data[0].FillTime, 10, 64)
	unixTimestamp := time.Unix(0, unixTimestampMillis*int64(time.Millisecond))
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	if unixTimestamp.Before(oneMinuteAgo) {
		sellTimeformat := unixTimestamp.Format("2006-01-02 15:04:05")
		logInfo := fmt.Sprintf("tradeFillsRsp.Data[0] fillTime exp: unixTimestamp.Before(oneMinuteAgo) %s", sellTimeformat)
		internal.PrintDebugLogToFile(logInfo)
		return internal.CALC_TRADE_PROFIT_DATA_SELL_TIME_ERROR
	}

	if tradeFillsRsp.Data[0].Side == "sell" && tradeFillsRsp.Data[1].Side == "buy" {
		sellProfit, _ := strconv.ParseFloat(tradeFillsRsp.Data[0].FillPnl, 10)
		sellFee, _ := strconv.ParseFloat(tradeFillsRsp.Data[0].Fee, 10)
		BuyFee, _ := strconv.ParseFloat(tradeFillsRsp.Data[1].Fee, 10)

		totalProfit := sellProfit + sellFee + BuyFee
		logTradeInfo := fmt.Sprintf("%s|CLOSE POSITION PROFIT: %f, %f, %f, %f", instId, totalProfit, sellProfit, sellFee, BuyFee)
		internal.PrintTradeLogToFile(logTradeInfo)

		return internal.RETURN_SUCCESS
	}

	logInfo := fmt.Sprintf("tradeFillsRsp.Data %s side exp: %s, %s", instId, tradeFillsRsp.Data[0].Side, tradeFillsRsp.Data[1].Side)
	internal.PrintDebugLogToFile(logInfo)
	return internal.CALC_TRADE_PROFIT_DATA_SIDE_TYPE_ERROR
}
