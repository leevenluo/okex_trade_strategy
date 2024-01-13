package main

import (
	"fmt"
	"time"
	"trade_strategy/internal"
	"trade_strategy/okerapi"
)

func main() {

	// var err error
	// var interval int64 = 10
	// firstRun := true

	// 初始化用户参数和http请求头
	req := internal.InitUserEnv("")

	// 初始化启动时间戳，为当天0点unix时间戳
	// var last_time int64
	// last_time, ret := internal.InitZeroTimestamp()
	// if ret != internal.RETURN_SUCCESS {
	// 	logInfo := fmt.Sprintf("internal.InitZeroTimestamp err: %s", ret)
	// 	internal.PrintDebugLogToFile(logInfo)
	// 	return
	// }

	// 可交易产品的信息列表
	instrument := make(map[string]*okerapi.InstrumentsDetail)
	okerapi.GetInstruments(req, instrument)

	// 针对已经建仓产品，定时轮询
	go okerapi.CheckAccountPositions()

	// 定时轮询符合条件的产品列表
	for true {
		// 获取所有资产产品信息
		productList, ret := okerapi.GetMarketTickers(req)
		if ret != internal.RETURN_SUCCESS {
			// 获取异常
			logInfo := fmt.Sprintf("okerapi.GetMarketTickers err: %s", ret)
			internal.PrintDebugLogToFile(logInfo)
			return
		}

		// 遍历每个产品确认交易逻辑
		for _, eachInstId := range productList {
			time.Sleep(1 * time.Second)
			// 获取产品对应的K线
			getCandlesResult, ret := okerapi.GetCandles(eachInstId.InstId)
			if ret != internal.RETURN_SUCCESS {
				logInfo := fmt.Sprintf("okerapi.GetCandles error: %s", ret)
				internal.PrintDebugLogToFile(logInfo)
				return
			}

			// 判断开仓条件
			status := okerapi.JudgeOrderCondition(*getCandlesResult)
			if status {
				// 开仓

				// 已经开仓过的产品，平仓前不再开仓，以及平仓后4H周期内也不再开仓
				tradeCondition := okerapi.JudgeTradeCondition(getCandlesResult.InstId)
				if !tradeCondition {
					logInfo := fmt.Sprintf("okerapi.JudgeTradeCondition err: %s", ret)
					internal.PrintDebugLogToFile(logInfo)
					continue
				}

				// 开仓数量计算
				_ = okerapi.CalcOrderCounts(*getCandlesResult)
				ordId, ret := okerapi.TradeOrder(eachInstId.InstId, "1", "100")
				if ret != internal.RETURN_SUCCESS {
					// 下单异常
					logInfo := fmt.Sprintf("okerapi.TradeOrder err: %s", ret)
					internal.PrintDebugLogToFile(logInfo)
					continue
				}

				// 下单成功，继续提交一笔止损订单，先拉取下单成交价，再计算止损位置后提交
				orderAvgPx, ret := okerapi.GetOrderInfoOfAvgPx(eachInstId.InstId, ordId)
				if ret != internal.RETURN_SUCCESS {
					// 拉取下单信息异常
					logInfo := fmt.Sprintf("okerapi.GetOrderInfoOfAvgPx err: %s", ret)
					internal.PrintDebugLogToFile(logInfo)
					continue
				}

				orderSLPx := okerapi.CalcSLOrderPx(*getCandlesResult, orderAvgPx)
				ret = okerapi.TradeSLOrder(eachInstId.InstId, orderSLPx, "100")
				if ret != internal.RETURN_SUCCESS {
					// 止损单下单信息异常
					logInfo := fmt.Sprintf("okerapi.TradeSLOrder err: %s", ret)
					internal.PrintDebugLogToFile(logInfo)
					continue
				}

				logInfo := fmt.Sprintf("%s order: %t", eachInstId.InstId, status)
				internal.PrintDebugLogToFile(logInfo)
			} else {
				// 不开仓
				logInfo := fmt.Sprintf("%s order: %t", eachInstId.InstId, status)
				internal.PrintDebugLogToFile(logInfo)
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
