package okerapi

import (
	"fmt"
	"math"
	"strconv"
	"trade_strategy/internal"
)

func JudgeOrderCondition(getCandlesResult GetCandlesResult) bool {

	h, _ := strconv.ParseFloat(getCandlesResult.Get4HCandlesWrapRsp.Data[0].HighPrice, 10)
	l, _ := strconv.ParseFloat(getCandlesResult.Get4HCandlesWrapRsp.Data[0].LowPrice, 10)

	c, _ := strconv.ParseFloat(getCandlesResult.Get15mCandlesWrapRsp.Data[0].ClosePrice, 10)
	o, _ := strconv.ParseFloat(getCandlesResult.Get15mCandlesWrapRsp.Data[0].OpenPrice, 10)

	K := h - l
	D := c - o

	n1 := 0.7

	logInfo := fmt.Sprintf("K: %f, D: %f, K*n1: %f, D >= K*n1: %t", K, D, K*n1, D >= K*n1)
	internal.PrintDebugLogToFile(logInfo)

	if D >= K*n1 {
		return true
	}

	return false
}

func CalcOrderCounts(getCandlesResult GetCandlesResult) string {
	var orderCounts string

	h, _ := strconv.ParseFloat(getCandlesResult.Get4HCandlesWrapRsp.Data[0].HighPrice, 10)
	l, _ := strconv.ParseFloat(getCandlesResult.Get4HCandlesWrapRsp.Data[0].LowPrice, 10)

	c, _ := strconv.ParseFloat(getCandlesResult.Get15mCandlesWrapRsp.Data[0].ClosePrice, 10)

	K := h - l

	n2 := 1.4

	counts := 50 / (K * n2 / c)

	// 向下取整
	downNum := math.Floor(counts)

	// 转换为字符串
	orderCounts = strconv.FormatFloat(downNum, 'f', 0, 64)

	logInfo := fmt.Sprintf("K: %f, counts: %f, orderCounts: %s", K, counts, orderCounts)
	internal.PrintDebugLogToFile(logInfo)

	return orderCounts
}

func CalcSLOrderPx(getCandlesResult GetCandlesResult, orderAvgPx string) string {
	var orderSLPx string

	h, _ := strconv.ParseFloat(getCandlesResult.Get4HCandlesWrapRsp.Data[0].HighPrice, 10)
	l, _ := strconv.ParseFloat(getCandlesResult.Get4HCandlesWrapRsp.Data[0].LowPrice, 10)

	P1, _ := strconv.ParseFloat(orderAvgPx, 10)

	K := h - l

	n2 := 1.4

	orderSLPx = strconv.FormatFloat(P1-K*n2, 'f', 0, 64)

	logInfo := fmt.Sprintf("K: %f, orderSLPx: %f, orderSLPx: %s", K, P1-K*n2, orderSLPx)
	internal.PrintDebugLogToFile(logInfo)
	return orderSLPx
}

// 判断止损价格是否调整
func JudgeSLPrice(marketTicker MarketTicker, currentAvgpx string, get4HCandlesWrapRsp GetCandlesWrapRsp) (bool, string) {

	h, _ := strconv.ParseFloat(get4HCandlesWrapRsp.Data[0].HighPrice, 10)
	l, _ := strconv.ParseFloat(get4HCandlesWrapRsp.Data[0].LowPrice, 10)
	K := h - l

	P2, _ := strconv.ParseFloat(marketTicker.Last, 10)
	P1, _ := strconv.ParseFloat(currentAvgpx, 10)

	logInfo := fmt.Sprintf("JudgeSLPrice, K: %f, P1: %f, P2: %f, P2 <= P1+2*K: %t", K, P1, P2, P2 <= P1+2*K)
	internal.PrintDebugLogToFile(logInfo)

	var newSLPrice string
	if P2 >= P1+2*K {
		newSLPrice = strconv.FormatFloat(P1+0.2*K, 'f', 0, 64)
		return true, newSLPrice
	}

	return false, newSLPrice
}

// 判断出场平仓条件是否满足
func JudgeClosePosition(get15mCandlesWrapBatchRsp GetCandlesWrapRsp) bool {
	// 取出最近的15m,计算D
	lastCandlesRspData := get15mCandlesWrapBatchRsp.Data[0]
	c, _ := strconv.ParseFloat(lastCandlesRspData.ClosePrice, 10)
	o, _ := strconv.ParseFloat(lastCandlesRspData.OpenPrice, 10)

	D := c - o

	// 计算近20个ATP20
	var totalHSum, totalLSum float64
	for index := 1; index < len(get15mCandlesWrapBatchRsp.Data); index++ {
		candlesRspData := get15mCandlesWrapBatchRsp.Data[index]
		h, _ := strconv.ParseFloat(candlesRspData.HighPrice, 10)
		l, _ := strconv.ParseFloat(candlesRspData.LowPrice, 10)
		totalHSum += h
		totalLSum += l
	}

	ATP20 := (totalHSum - totalLSum) / 20

	n3 := 0.2

	logInfo := fmt.Sprintf("JudgeClosePosition, c: %f, o: %f, D: %f, ATP20: %f, math.Abs(D) >= ATP20*n3: %t", c, o, D, ATP20, math.Abs(D) >= ATP20*n3)
	internal.PrintDebugLogToFile(logInfo)

	// if math.Abs(D) >= ATP20*n3 {
	if D < 0 && math.Abs(D) >= ATP20*n3 {
		return true
	}

	return false
}
