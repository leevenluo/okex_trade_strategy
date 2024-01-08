package okerapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"trade_strategy/internal"
)

// InstrumentsDetail 可交易产品的信息列表struct定义
type InstrumentsDetail struct {
	InstType  string `json:"instType"`
	InstId    string `json:"instId"`
	Uly       string `json:"uly"`
	Category  string `json:"category"`
	BaseCcy   string `json:"baseCcy"`
	QuoteCcy  string `json:"quoteCcy"`
	SettleCcy string `json:"settleCcy"`
	CtVal     string `json:"ctVal"`
	CtMult    string `json:"ctMult"`
	CtValCcy  string `json:"ctValCcy"`
	OptType   string `json:"optType"`
	Stk       string `json:"stk"`
	ListTime  string `json:"listTime"`
	ExpTime   string `json:"expTime"`
	Lever     string `json:"lever"`
	TickSz    string `json:"tickSz"`
	LotSz     string `json:"lotSz"`
	MinSz     string `json:"minSz"`
	CtType    string `json:"ctType"`
	Alias     string `json:"alias"`
	State     string `json:"state"`
}

type InstrumentsInfo struct {
	Code string              `json:"code"`
	Msg  string              `json:"msg"`
	Data []InstrumentsDetail `json:"data"`
}

// GetInstruments 可交易产品的信息列表struct定义
func GetInstruments(req *http.Request, instrument map[string]*InstrumentsDetail) string {
	req.URL.Path = "/api/v5/public/instruments"
	inst_type := []string{"SWAP"}
	for i := 0; i < len(inst_type); i++ {
		req.URL.RawQuery = ""
		request_url := req.URL.Query()
		request_url.Add("instType", inst_type[i])
		req.URL.RawQuery = request_url.Encode()
		fmt.Println(req.URL.String())

		client := &http.Client{}
		resp, _ := client.Do(req)

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(body))

		var instrumentInfo InstrumentsInfo
		err := json.Unmarshal(body, &instrumentInfo)
		if err != nil {
			logInfo := fmt.Sprintf("GetInstruments Unmarshal err: %s", err.Error())
			internal.PrintDebugLogToFile(logInfo)
			return internal.GET_INSTRUMENTS_UNMARSHAL_RSP_ERROR
		}

		if instrumentInfo.Code != internal.RETURN_SUCCESS {
			logInfo := fmt.Sprintf("orderAlgoRsp err, code: %s, msg: %s", instrumentInfo.Code, instrumentInfo.Msg)
			internal.PrintDebugLogToFile(logInfo)
			return internal.GET_INSTRUMENTS_POST_RSP_ERROR
		}

		for _, each_item := range instrumentInfo.Data {
			instrument[each_item.InstId] = &each_item
		}

		logInfo := fmt.Sprintf("GetInstruments %s: %d", inst_type[i], len(instrument))
		internal.PrintDebugLogToFile(logInfo)
	}

	return internal.RETURN_SUCCESS
}
