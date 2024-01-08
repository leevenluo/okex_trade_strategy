package internal

import (
	"fmt"
	"strconv"
	"time"
)

type TickerDetailsUnit struct {
	InstId string
	LastFundingRate string // 最近1次资金费率
	LastFundingRateTime string // 最近1次资金费率更新时间
	LastPriceGap string // 最近1次价差
	LastPriceGapTime string // 最近1次价差更新时间
}

type TickersDetails struct {
	TickerUnit TickerDetailsUnit
	Last3FundingRate []string // 最近3次资金费率，8小时不变，所以应该是近24小时的周期
	Last3PriceGap []string // 最近3次价差，基本实时再变，可考虑拉长
}

type TickersStatus struct {
	// key instd swap id
	TickersInfo map[string]*TickersDetails
}

func UpdateTickerFundingRate(
	ticker_status *TickersStatus,
	c_ticker_unit chan TickerDetailsUnit) {
	t_unit := <- c_ticker_unit
	ticker_info, ok := ticker_status.TickersInfo[t_unit.InstId]
	if ok {
		ticker_info.TickerUnit = t_unit
	} else {
		var t_info TickersDetails
		t_info.TickerUnit = t_unit
		ticker_status.TickersInfo[t_unit.InstId] = &t_info
	}
}

func CheckValidTickerForBill(
	ticker_status *TickersStatus) (err error) {
	for true {
		time.Sleep(2000 * time.Millisecond)
		ticker_status_copy := *ticker_status

		for ticker_id := range ticker_status_copy.TickersInfo {
			last_funding_rate := ticker_status_copy.TickersInfo[ticker_id].TickerUnit.LastFundingRate
			last_price_gap := ticker_status_copy.TickersInfo[ticker_id].TickerUnit.LastPriceGap

			var last_funding_rate_f, last_price_gap_f float64
			last_funding_rate_f, err = strconv.ParseFloat(last_funding_rate, 32)
			last_price_gap_f, err = strconv.ParseFloat(last_price_gap, 32)
			if (last_funding_rate_f >= 0.0015) && (last_price_gap_f >= 0.005) {
				// 下单
				fmt.Println("true,", ticker_id, last_funding_rate, last_price_gap)
			} else {
				// 不下单
				fmt.Println("false,", ticker_id, last_funding_rate, last_price_gap)
				//_, ok := instrument[ticker_id]
				//if ok {
				//	result := instrument[ticker_id].TickSz + ", "
				//	result += instrument[ticker_id].LotSz + ", "
				//	result += instrument[ticker_id].MinSz + ", "
				//	fmt.Println(result)
				//}
			}
		}
	}

	return err
}