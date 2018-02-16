package main

import (
	"context"
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"net/http"
	"regexp"
	"net/url"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	// @BotFather gives you this
	BotToken   = "417201087:AAHGtsGj99alk2B2YXAz8x3otYZoB0mEQuE"
	WebhookURL = "https://a272990b.ngrok.io"
	URLPATH = "https://54272f50.ngrok.io"
)

type ErrorType struct {
	Error string `json:"error"`
}

func getInfo(userID int64, bot *tgbotapi.BotAPI) {
	type Record struct {
		Id int `json:"id"`
		Ticker string `json:"ticker"`
		Vol int `json:"vol"`
	}
	type Response struct {
		Balance float32 `json:"balance"`
		Records []Record `json:"records"`
	}
	type GetInfoResponse struct {
		GotResponse Response `json:"response"`
	}
	msg := ""
	urlstr := URLPATH
	urlstr += "/getinfo?"
	v := url.Values{}
	v.Set("u_id", strconv.FormatInt(userID, 10))
	urlstr += v.Encode()
	//msg = urlstr
	resp, _ := http.Get(urlstr)
	respbody, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(respbody))
	getInfoResponse := &GetInfoResponse{}
	if resp.StatusCode == http.StatusBadRequest {
		newErr := &ErrorType{}
		json.Unmarshal(respbody, newErr)
		msg += "Ошибка: " + newErr.Error
	} else {
		json.Unmarshal(respbody, getInfoResponse)
		msg += "Баланс: " + strconv.FormatFloat(float64(getInfoResponse.GotResponse.Balance), 'f', 2, 64) + "\n"
		for _, record := range getInfoResponse.GotResponse.Records {
			msg += "ID: " + strconv.Itoa(record.Id) + "\n"
			msg += "Инструмент: " + record.Ticker + "\n"
			msg += "Количество: " + strconv.Itoa(record.Vol) + "\n"
			if len(getInfoResponse.GotResponse.Records) > 1 {
				msg += "\n"
			}
		}
	}
	bot.Send(tgbotapi.NewMessage(
		userID,
		msg,
	))
}

func initialMessage(userID int64, bot *tgbotapi.BotAPI) {
	bot.Send(tgbotapi.NewMessage(
		userID,
		"Здравствуйте! Список доступных команд:\n/newuser <balance> - Создание нового пользователя с балансом balance\n/getinfo - Посмотреть свои позиции и баланс\n/gethistory - Посмотреть историю своих операций\n/buyreq <ticker> <volume> <price>- Запрос на покупку ticker в количетве volume по цене price\n/sellreq <ticker> <volume> <price>- Запрос на продажу ticker в количестве volume по цене price\n/abortreq <req_id> - Запрос на отмену операции с id req_id\n/tradehistory - Посмотреть последнюю историю торгов\n/getrequests - Посмотреть свои запросы",
	))
}

func newUser(userID int64, balance string, bot *tgbotapi.BotAPI) {
	type Response struct {
		Id int `json:"id"`
	}
	type NewUserResponse struct {
		GotResponse Response `json:"response"`
	}
	var msg string
	msg = ""
	urlstr := URLPATH
	urlstr += "/newuser?"
	v := url.Values{}
	v.Set("u_id", strconv.FormatInt(userID, 10))
	v.Add("cash", balance)
	urlstr += v.Encode()
	resp, _ := http.Get(urlstr)
	//log.Println(resp.Body)
	respbody, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(respbody))
	newUserResponse := &NewUserResponse{}
	if resp.StatusCode == http.StatusBadRequest {
		newErr := &ErrorType{}
		json.Unmarshal(respbody, newErr)
		msg += "Ошибка: " + newErr.Error
	} else {
		json.Unmarshal(respbody, newUserResponse)
		msg += "Пользователь создан"
	}
	//msg += urlstr
	bot.Send(tgbotapi.NewMessage(
		userID,
		msg,
	))
}

func getHistory(isMy bool, userID int64, bot *tgbotapi.BotAPI) {
	type Record struct {
		Id int `json: "id"`
		Ticker string `json: "ticker"`
		Is_buy int `json: "is_buy"`
		Price float32 `json: "price"`
		Vol int `json: "vol"`
	}
	type Response struct {
		Records []Record `json: "records"`
	}
	type GetInfoResponse struct {
		GotResponse Response `json: "response"`
	}
	var msg string
	msg = ""
	urlstr := URLPATH
	if isMy {
		urlstr += "/gethistory?"
		v := url.Values{}
		v.Set("u_id", strconv.FormatInt(userID, 10))
		urlstr += v.Encode()
	} else {
		urlstr += "/tradehistory"
	}
	resp, _ := http.Get(urlstr)
	respbody, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(respbody))
	getInfoResponse := &GetInfoResponse{}
	if resp.StatusCode == http.StatusBadRequest {
		newErr := &ErrorType{}
		json.Unmarshal(respbody, newErr)
		msg += "Ошибка: " + newErr.Error
	} else {
		json.Unmarshal(respbody, getInfoResponse)
		for _, record := range getInfoResponse.GotResponse.Records {
			msg += "ID: " + strconv.Itoa(record.Id) + "\n"
			msg += "Инструмент: " + record.Ticker + "\n"
			if record.Is_buy == 1 {
				msg += "Тип: покупка\n"
			} else {
				msg += "Тип: продажа\n"
			}
			msg += "Цена: " + strconv.FormatFloat(float64(record.Price), 'f', 2, 64) + "\n"
			msg += "Количество: " + strconv.Itoa(record.Vol) + "\n"
			if len(getInfoResponse.GotResponse.Records) > 1 {
				msg += "\n"
			}
		}
	}
	//msg += urlstr
	bot.Send(tgbotapi.NewMessage(
		userID,
		msg,
	))
}

func sendRequest(typeOp string, userID int64,  params []string, bot *tgbotapi.BotAPI) {
	type Response struct {
		Id int `json:"id"`
	}
	type GetReqResponse struct {
		GotResponse Response `json:"response"`
	}
	var msg string
	msg = ""
	urlstr := URLPATH
	urlstr += "/" + typeOp + "req?"
	v := url.Values{}
	v.Set("u_id", strconv.FormatInt(userID, 10))
	v.Add("ticker", params[0])
	v.Add("vol", params[1])
	v.Add("price", params[2])
	urlstr += v.Encode()
	resp, _ := http.Get(urlstr)
	respbody, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(respbody))
	getReqResponse := &GetReqResponse{}
	if resp.StatusCode == http.StatusBadRequest {
		newErr := &ErrorType{}
		json.Unmarshal(respbody, newErr)
		msg += "Ошибка: " + newErr.Error
	} else {
		json.Unmarshal(respbody, getReqResponse)
		msg += "Операция отправлена. ID операции: " + strconv.Itoa(getReqResponse.GotResponse.Id)
	}
	//msg += urlstr
	bot.Send(tgbotapi.NewMessage(
		userID,
		msg,
	))
}

func sendCancelRequest(userID int64, reqID string, bot *tgbotapi.BotAPI) {
	type Response struct {
		Id int `json: "id"`
	}
	type GetCancelResponse struct {
		GotResponse Response `json: "response"`
	}
	var msg string
	msg = ""
	urlstr := URLPATH
	urlstr += "/abortreq?"
	v := url.Values{}
	v.Set("u_id", strconv.FormatInt(userID, 10))
	v.Add("req_id", reqID)
	urlstr += v.Encode()
	resp, _ := http.Get(urlstr)
	respbody, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(respbody))
	getCancelResponse := &GetCancelResponse{}
	if resp.StatusCode == http.StatusBadRequest {
		newErr := &ErrorType{}
		json.Unmarshal(respbody, newErr)
		msg += "Ошибка: " + newErr.Error
	} else {
		json.Unmarshal(respbody, getCancelResponse)
		msg += "Операция отменена."
	}
	//msg += urlstr
	bot.Send(tgbotapi.NewMessage(
		userID,
		msg,
	))
}

func getRequests(userID int64, bot *tgbotapi.BotAPI) {
	type Request struct {
		ID     int `json:"id"`
		UserID int `json:"u_id"`
		Ticker string `json:"ticker"`
		Vol    int `json:"vol"`
		Price  float32 `json:"price"`
		IsBuy  int `json:"is_buy"`
	}
	type Response struct {
		Requests []Request `json:"records"`
	}
	type GetReqsResponse struct {
		GotResponse Response `json:"response"`
	}
	var msg string
	msg = ""
	urlstr := URLPATH
	urlstr += "/getrequests?"
	v := url.Values{}
	v.Set("u_id", strconv.FormatInt(userID, 10))
	urlstr += v.Encode()
	resp, _ := http.Get(urlstr)
	respbody, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(respbody))
	getReqsResponse := &GetReqsResponse{}
	if resp.StatusCode == http.StatusBadRequest {
		newErr := &ErrorType{}
		json.Unmarshal(respbody, newErr)
		msg += "Ошибка: " + newErr.Error
	} else {
		json.Unmarshal(respbody, getReqsResponse)
		fmt.Printf("%#v\n", getReqsResponse)
		for _, record := range getReqsResponse.GotResponse.Requests {
			msg += "ID: " + strconv.Itoa(record.ID) + "\n"
			msg += "Инструмент: " + record.Ticker + "\n"
			if record.IsBuy == 1 {
				msg += "Тип: покупка\n"
			} else {
				msg += "Тип: продажа\n"
			}
			msg += "Цена: " + strconv.FormatFloat(float64(record.Price), 'f', 2, 64) + "\n"
			msg += "Количество: " + strconv.Itoa(record.Vol) + "\n"
			if len(getReqsResponse.GotResponse.Requests) > 1 {
				msg += "\n"
			}
		}
	}
	//msg += urlstr
	bot.Send(tgbotapi.NewMessage(
		userID,
		msg,
	))
}

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)
	
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}
	
	updates := bot.ListenForWebhook("/")
	
	go http.ListenAndServe(":8080", nil)
	fmt.Println("start listen :8080")

	for update := range updates {
		regexTask := regexp.MustCompile(`^\/([a-z]+)\s?(SPFB.+?)?\s?(\d+)?\s?(\d+)?$`)
		if regexTask.MatchString(update.Message.Text) {
			regexResult := regexTask.FindStringSubmatch(update.Message.Text)
			log.Println(regexResult)	
			switch (regexResult[1]){
			case "start":
				initialMessage(update.Message.Chat.ID, bot)
			case "getinfo":
				getInfo(update.Message.Chat.ID, bot)
			case "newuser":
				newUser(update.Message.Chat.ID, regexResult[3], bot)
			case "tradehistory":
				getHistory(false, update.Message.Chat.ID, bot)
			case "gethistory":
				getHistory(true, update.Message.Chat.ID, bot)
			case "buyreq":
				sendRequest("buy", update.Message.Chat.ID, regexResult[2:], bot)
			case "sellreq":
				sendRequest("sell", update.Message.Chat.ID, regexResult[2:], bot)
			case "abortreq":
				sendCancelRequest(update.Message.Chat.ID, regexResult[3], bot)
			case "getrequests":
				getRequests(update.Message.Chat.ID, bot)
			default:
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Неизвестная команда. Для просмотра возможных команд введите /start",
				))
			}
		} else {
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Неверный формат ввода. Для просмотра формата ввода введите /start",
			))
		}
	}
	return nil
}


func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
