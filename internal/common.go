package internal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	user_secret_key = "90724260537048306A14586DC7271E88"
	//user_secret_key = "6D60CCDDF466B4ACDA6D6B382A508209"
	user_api_key = "ee0eb155-3f66-420d-850a-c21c68ba5110"
	//user_api_key = "bdab5985-20a8-4d47-9023-da3b8abb8c18"
	user_passphrase = "testtest"

	root_path = "https://www.okex.com"

	ROOT_PATH = "https://www.okex.com"

	RETURN_SUCCESS = "0"

	// Order Return Code
	// getOrderInfoOfAvgPx
	GET_ORDER_INFO_UNMARSHAL_RSP_ERROR   = "100"
	GET_ORDER_INFO_POST_RSP_ERROR        = "101"
	GET_ORDER_INFO_POST_DATA_ERROR       = "102"
	GET_ORDER_INFO_POST_DATA_STATE_ERROR = "102"

	// TradeOrder
	TRADE_ORDER_REQ_ERROR            = "110"
	TRADE_ORDER_UNMARSHAL_RSP_ERROR  = "111"
	TRADE_ORDER_POST_RSP_ERROR       = "112"
	TRADE_ORDER_POST_DATA_ERROR      = "113"
	TRADE_ORDER_POST_DATA_CODE_ERROR = "114"

	// TradeSLOrder
	TRADE_SL_ORDER_UNMARSHAL_RSP_ERROR = "120"
	TRADE_SL_ORDER_POST_RSP_ERROR      = "121"
	TRADE_SL_ORDER_POST_DATA_ERROR     = "122"

	// InitZeroTimestamp
	INIT_TIME_PARSE_IN_LOCATION_ERROR = "130"

	// GetMarketTickers
	GET_MARKET_TICKERS_UNMARSHAL_RSP_ERROR = "140"

	// GetCandles
	GET_CANDLES_RESULT_EMPTY = "150"
	GET_CANDLES_RESULT_ERROR = "151"

	// CheckAdjustSLPrice
	CHECK_ADJUST_SLPRICE_KLINE_ERROR = "160"

	// getPeriodCandles
	GET_PERIOD_CANDLES_UNMARSHAL_RSP_ERROR = "170"
	GET_PERIOD_CANDLES_POST_RSP_ERROR      = "171"
	GET_PERIOD_CANDLES_POST_DATA_ERROR     = "172"

	// GetMarketTicker
	GET_MARKET_TICKER_UNMARSHAL_RSP_ERROR = "180"
	GET_MARKET_TICKER_POST_RSP_ERROR      = "181"
	GET_MARKET_TICKER_POST_DATA_ERROR     = "182"

	// AdjustSLPrice
	ADJUST_SL_RPICE_UNMARSHAL_RSP_ERROR = "190"
	ADJUST_SL_RPICE_POST_RSP_ERROR      = "191"
	ADJUST_SL_RPICE_POST_DATA_ERROR     = "192"
	ADJUST_SL_RPICE_RESULT_ERROR        = "193"

	// ClosePosition
	CLOSE_POSITION_UNMARSHAL_RSP_ERROR = "200"
	CLOSE_POSITION_POST_RSP_ERROR      = "201"
	CLOSE_POSITION_POST_DATA_ERROR     = "202"

	// GetInstruments
	GET_INSTRUMENTS_UNMARSHAL_RSP_ERROR = "210"
	GET_INSTRUMENTS_POST_RSP_ERROR      = "211"
)

// InitUserEnv 初始化用户访问参数
// accounts_path, 需要验证的path参数，比如/api/v5/account/account-position-risk
func InitUserEnv(accounts_path string) *http.Request {
	var req *http.Request
	now_time := time.Now().UTC().Format(time.RFC3339)
	crypto_content := now_time + "GET" + accounts_path
	sha256_crypoto := hmac.New(sha256.New, []byte(user_secret_key))
	sha256_crypoto.Write([]byte(crypto_content))
	user_sign := base64.StdEncoding.EncodeToString([]byte(sha256_crypoto.Sum(nil)))
	//fmt.Println("sign: " + user_sign)
	//fmt.Println("now_time: " + now_time)

	req, _ = http.NewRequest("GET", root_path, nil)
	req.Header.Add("OK-ACCESS-KEY", user_api_key)
	req.Header.Add("OK-ACCESS-SIGN", user_sign)
	req.Header.Add("OK-ACCESS-TIMESTAMP", now_time)
	req.Header.Add("OK-ACCESS-PASSPHRASE", user_passphrase)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-simulated-trading", "1")

	//fmt.Printf("Request Header: %v\n", req.Header)
	return req
}

// InitUserEnvPost 初始化用户访问参数
// accounts_path, 需要验证的path参数，比如/api/v5/account/account-position-risk
func InitUserEnvPost(accounts_path string, body string) (req *http.Request, err error) {
	now_time := time.Now().UTC().Format(time.RFC3339)
	crypto_content := now_time + "POST" + accounts_path + body
	sha256_crypoto := hmac.New(sha256.New, []byte(user_secret_key))
	sha256_crypoto.Write([]byte(crypto_content))
	user_sign := base64.StdEncoding.EncodeToString([]byte(sha256_crypoto.Sum(nil)))
	//fmt.Println("sign: " + sign)
	//fmt.Println(crypto_content)

	req, _ = http.NewRequest("POST", root_path, strings.NewReader(body))
	req.Header.Add("OK-ACCESS-KEY", user_api_key)
	req.Header.Add("OK-ACCESS-SIGN", user_sign)
	req.Header.Add("OK-ACCESS-TIMESTAMP", now_time)
	req.Header.Add("OK-ACCESS-PASSPHRASE", user_passphrase)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-simulated-trading", "1")
	//req.Body.Read(strings.NewReader(now_time))

	//fmt.Printf("Request Header: %v\n", req.Header)
	err = nil
	return
}

// InitZeroTimestamp 时间间隔判断函数
// interval 间隔时间，单位s(秒), 比如60就是1分钟,
func InitZeroTimestamp() (int64, string) {
	var zero_timestamp int64
	now_time := time.Now().Format("2006-01-02 15:04:05")
	now_time_array := strings.Split(now_time, " ")
	now_time_zero := now_time_array[0] + " 00:00:00"
	zero_timestamp_time, err := time.ParseInLocation("2006-01-02 15:04:05", now_time_zero, time.Local)

	if err != nil {
		logInfo := fmt.Sprintf("InitZeroTimestamp time.ParseInLocation err: %s", err.Error())
		PrintDebugLogToFile(logInfo)
		return zero_timestamp, INIT_TIME_PARSE_IN_LOCATION_ERROR
	}

	zero_timestamp = zero_timestamp_time.Unix()
	return zero_timestamp, RETURN_SUCCESS
}

// TimeIntervalSuccess 时间间隔外判断函数，use_type=1，间隔内true，use_type=2，间隔外true
// 例子：间隔是60s，use_type=1，间隔60s内返回都是true，use_type=2，间隔60s外返回都是true
// interval 间隔时间，单位s(秒), 比如60就是1分钟,
func TimeIntervalSuccess(last_time *int64, interval int64, use_type int) (status bool, err error) {
	// 当前日期unix时间戳
	current_timestamp := time.Now().Unix()

	//fmt.Println(current_timestamp, *last_time, current_timestamp - *last_time)
	//fmt.Println(interval, (current_timestamp - *last_time) > interval)

	// 计算时间间隔
	if (use_type == 1) && ((current_timestamp - *last_time) < interval) {
		return true, err
	} else if (use_type == 2) && ((current_timestamp - *last_time) >= interval) {
		*last_time = current_timestamp
		return true, err
	} else {
		return false, err
	}
}

func PrintDebugLogToFile(log string) {
	t := time.Now()
	currentTime := fmt.Sprintf("%d-%d-%d %d:%d:%d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

	// 检查目录是否存在，不存在则创建
	dirPath := "/home/lighthouse/leeven/okex_trade_strategy/log/"
	_, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(dirPath, os.ModePerm)
			if err != nil {
				fmt.Println("create log dir error", err.Error())
				return
			}
		} else {
			fmt.Println("check log dir error", err.Error())
			return
		}
	}

	// 打开文件,文件名前缀为时间戳
	today := time.Now().Format("2006-01-02")
	fileName := dirPath + fmt.Sprintf("%s_%s.log", today, "DEBUG")
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("open file error", err.Error())
		return
	}
	defer f.Close()

	// 在日志前面加上文件名和行号
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		// file绝对路径截取最后一个/后面的内容
		file = file[strings.LastIndex(file, "/")+1:]
		log = fmt.Sprintf("%s|%s|%s|%d|%s|%s\n",
			currentTime, runtime.FuncForPC(pc).Name(), file, line, "DEBUG", log)
	} else {
		fmt.Println("runtime.Caller error")
	}

	// 写入文件
	_, err = f.WriteString(log)
	if err != nil {
		fmt.Println("write file error", err.Error())
		return
	}
}
