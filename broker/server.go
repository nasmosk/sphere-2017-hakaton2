package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sphere-2017-2-hakaton2/database"
	"strconv"
)

var reqCh = make(chan database.Request)

func PrepareTestData(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS clients;`,
		`DROP TABLE IF EXISTS positions;`,
		`DROP TABLE IF EXISTS orders_history;`,
		`DROP TABLE IF EXISTS request;`,
		`DROP TABLE IF EXISTS stat;`,

		`CREATE TABLE clients (
id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
login_id int NOT NULL,
-- password varchar(300) NOT NULL,
balance int NOT NULL
)`,
		`CREATE TABLE positions (
    id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id int NOT NULL,
    ticker varchar(300) NOT NULL,
    vol int NOT NULL,
    KEY user_id(user_id)
);`,
		`CREATE TABLE orders_history (
    id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    time int NOT NULL,
    user_id int,
    ticker varchar(300) NOT NULL,
    vol int NOT NULL,
    price float not null,
    is_buy int not null,
    KEY user_id(user_id)
);`,
		`CREATE TABLE request ( -- запросы
    id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id int,
    ticker varchar(300) NOT NULL,
    vol int NOT NULL,
    price float NOT NULL,
    is_buy int not null, -- 1 - покупаем, 0 - продаем
    KEY user_id(user_id)
);`,
		`CREATE TABLE stat ( 
    id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    time int,` + "`interval` int," + `
    open float,
    high float,
    low float,
    close float,
    volume int,
    ticker varchar(300),
    KEY id(id)
);`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func sendErr(w http.ResponseWriter, r *http.Request, err string, status int) {
	logmsg := "	error : " + err
	fmt.Println(logmsg)
	response := make(map[string]interface{})
	response["error"] = err
	fnl, _ := json.Marshal(response)
	w.WriteHeader(status)
	io.WriteString(w, string(fnl))
}

func sendId(w http.ResponseWriter, r *http.Request, id int) {
	records := make(map[string]interface{})
	records["id"] = id
	response := make(map[string]interface{})
	response["response"] = records
	fnl, _ := json.Marshal(response)
	io.WriteString(w, string(fnl))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	sendErr(w, r, "wrong method", http.StatusNotFound)
	return
}

// 127.0.0.1:8081/newuser?u_id=228&cash=1000
func newuserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	//check u_id
	if len(r.FormValue("u_id")) == 0 {
		sendErr(w, r, "no parm u_id", http.StatusBadRequest)
		return
	}
	UserID, err := strconv.Atoi(r.FormValue("u_id"))
	if err != nil {
		sendErr(w, r, "wrong param u_id", http.StatusBadRequest)
		return
	}

	//check cash
	if len(r.FormValue("cash")) == 0 {
		sendErr(w, r, "no parm cash", http.StatusBadRequest)
		return
	}
	UserCash, err := strconv.Atoi(r.FormValue("cash"))
	if err != nil {
		sendErr(w, r, "wrong param cash", http.StatusBadRequest)
		return
	}

	fmt.Println("	try create user " + strconv.Itoa(UserID) + " with cash " + strconv.Itoa(UserCash))
	//CREATE user and return id

	id, err := database.CreateClient(UserID, UserCash)

	sendId(w, r, id)
}

// http://127.0.0.1:8081/getinfo?u_id=123
func getinfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	//check u_id
	if len(r.FormValue("u_id")) == 0 {
		sendErr(w, r, "no parm u_id", http.StatusBadRequest)
		return
	}
	UserID, err := strconv.Atoi(r.FormValue("u_id"))
	if err != nil {
		sendErr(w, r, "wrong param u_id", http.StatusBadRequest)
		return
	}

	fmt.Println("	request info for user_id " + strconv.Itoa(UserID))

	user, err := database.GetClientByLoginID(UserID)
	items := database.GetPositions(user.ID)
	if err != nil {
		sendErr(w, r, err.Error(), http.StatusBadRequest)

	}

	records := make(map[string]interface{})
	records["balance"] = user.Balance
	records["records"] = items

	response := make(map[string]interface{})
	response["response"] = records

	fnl, _ := json.Marshal(response)
	io.WriteString(w, string(fnl))
}

// http://127.0.0.1:8081/gethistory?u_id=123
func gethistoryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	//check u_id
	if len(r.FormValue("u_id")) == 0 {
		sendErr(w, r, "no parm u_id", http.StatusBadRequest)
		return
	}
	UserID, err := strconv.Atoi(r.FormValue("u_id"))
	if err != nil {
		sendErr(w, r, "wrong param u_id", http.StatusBadRequest)
		return
	}

	fmt.Println("	request history for user_id " + strconv.Itoa(UserID))

	user, err := database.GetClientByLoginID(UserID)
	items := database.GetOrderByID(user.ID)

	records := make(map[string]interface{})
	records["records"] = items

	response := make(map[string]interface{})
	response["response"] = records

	fnl, _ := json.Marshal(response)
	io.WriteString(w, string(fnl))

}

// http://127.0.0.1:8081/buyreq?u_id=228&cash=1000&ticker=qoqo.123&vol=20&price=100
func buyreqHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	//check u_id
	if len(r.FormValue("u_id")) == 0 {
		sendErr(w, r, "no parm u_id", http.StatusBadRequest)
		return
	}
	UserID, err := strconv.Atoi(r.FormValue("u_id"))
	if err != nil {
		sendErr(w, r, "wrong param u_id", http.StatusBadRequest)
		return
	}
	//check ticker
	if len(r.FormValue("ticker")) == 0 {
		sendErr(w, r, "no parm ticker", http.StatusBadRequest)
		return
	}
	Ticker := r.FormValue("ticker")
	//check vol
	if len(r.FormValue("vol")) == 0 {
		sendErr(w, r, "no parm vol", http.StatusBadRequest)
		return
	}
	Vol, err := strconv.Atoi(r.FormValue("vol"))
	if err != nil {
		sendErr(w, r, "wrong param vol", http.StatusBadRequest)
		return
	}
	//check price
	if len(r.FormValue("price")) == 0 {
		sendErr(w, r, "no parm price", http.StatusBadRequest)
		return
	}
	Price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		sendErr(w, r, "wrong param price", http.StatusBadRequest)
		return
	}
	fmt.Println("	request to buy for user_id:" + strconv.Itoa(UserID) + " ticker:" + Ticker + " vol:" + strconv.Itoa(Vol) + " price:" + strconv.FormatFloat(Price, 'f', 6, 64))
	//HERE buy and get id
	user, err := database.GetClientByLoginID(UserID)

	req := database.Request{
		UserID: user.ID,
		Ticker: Ticker,
		Vol:    Vol,
		Price:  float32(Price),
		IsBuy:  1,
	}
	reqCh <- req
	id, err := database.InsertRequest(req)

	sendId(w, r, id)
}

// 127.0.0.1:8081/sellreq?u_id=228&cash=1000&ticker=qoqo.123&vol=20&price=100
func sellreqHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	//check u_id
	if len(r.FormValue("u_id")) == 0 {
		sendErr(w, r, "no parm u_id", http.StatusBadRequest)
		return
	}
	UserID, err := strconv.Atoi(r.FormValue("u_id"))
	if err != nil {
		sendErr(w, r, "wrong param u_id", http.StatusBadRequest)
		return
	}
	//check ticker
	if len(r.FormValue("ticker")) == 0 {
		sendErr(w, r, "no parm ticker", http.StatusBadRequest)
		return
	}
	Ticker := r.FormValue("ticker")
	//check vol
	if len(r.FormValue("vol")) == 0 {
		sendErr(w, r, "no parm vol", http.StatusBadRequest)
		return
	}
	Vol, err := strconv.Atoi(r.FormValue("vol"))
	if err != nil {
		sendErr(w, r, "wrong param vol", http.StatusBadRequest)
		return
	}
	//check price
	if len(r.FormValue("price")) == 0 {
		sendErr(w, r, "no parm price", http.StatusBadRequest)
		return
	}
	Price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		sendErr(w, r, "wrong param price", http.StatusBadRequest)
		return
	}
	fmt.Println("	request to sell for user_id:" + strconv.Itoa(UserID) + " ticker:" + Ticker + " vol:" + strconv.Itoa(Vol) + " price:" + strconv.FormatFloat(Price, 'f', 6, 64))

	//HERE sell and get id
	user, err := database.GetClientByLoginID(UserID)

	req := database.Request{
		UserID: user.ID,
		Ticker: Ticker,
		Vol:    Vol,
		Price:  float32(Price),
		IsBuy:  1,
	}
	reqCh <- req

	id, err := database.InsertRequest(req)

	if err != nil {
		panic(err)
	}

	sendId(w, r, id)
}

// http://127.0.0.1:8081/abortreq?u_id=228&req_id=123
func abortreqHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	//check u_id
	if len(r.FormValue("u_id")) == 0 {
		sendErr(w, r, "no parm u_id", http.StatusBadRequest)
		return
	}
	UserID, err := strconv.Atoi(r.FormValue("u_id"))
	if err != nil {
		sendErr(w, r, "wrong param u_id", http.StatusBadRequest)
		return
	}
	//check req_id
	if len(r.FormValue("req_id")) == 0 {
		sendErr(w, r, "no parm req_id", http.StatusBadRequest)
		return
	}
	ReqID, err := strconv.Atoi(r.FormValue("req_id"))
	if err != nil {
		sendErr(w, r, "wrong param req_id", http.StatusBadRequest)
		return
	}

	fmt.Println("abort request for user_id:" + strconv.Itoa(UserID) + " requset id:" + strconv.Itoa(ReqID))

	user, err := database.GetClientByLoginID(UserID)

	err = database.DeleteRequest(user.ID, ReqID)
	if err != nil {
		panic(err)
	}

	sendId(w, r, ReqID)
}

// http://127.0.0.1:8081/tradehistory
func tradehistoryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	fmt.Println("	request trade history")

	items := database.GetStat()

	records := make(map[string]interface{})
	records["records"] = items

	response := make(map[string]interface{})
	response["response"] = records

	fnl, _ := json.Marshal(response)
	io.WriteString(w, string(fnl))
}

func getrequestsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	//check u_id
	if len(r.FormValue("u_id")) == 0 {
		sendErr(w, r, "no parm u_id", http.StatusBadRequest)
		return
	}
	UserID, err := strconv.Atoi(r.FormValue("u_id"))
	if err != nil {
		sendErr(w, r, "wrong param u_id", http.StatusBadRequest)
		return
	}

	fmt.Println("	request current requests for user_id " + strconv.Itoa(UserID))

	user, err := database.GetClientByLoginID(UserID)
	items := database.GetRequests(user.ID)

	records := make(map[string]interface{})
	records["records"] = items

	response := make(map[string]interface{})
	response["response"] = records

	fnl, _ := json.Marshal(response)
	io.WriteString(w, string(fnl))
}

func main() {
	db := database.InitDB()
	PrepareTestData(db)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/newuser", newuserHandler)
	http.HandleFunc("/getinfo", getinfoHandler)
	http.HandleFunc("/gethistory", gethistoryHandler)
	http.HandleFunc("/buyreq", buyreqHandler)
	http.HandleFunc("/sellreq", sellreqHandler)
	http.HandleFunc("/abortreq", abortreqHandler)
	http.HandleFunc("/tradehistory", tradehistoryHandler)
	http.HandleFunc("/getrequests", getrequestsHandler)

	// go grcp_client.GrpcClient(reqCh)

	fmt.Println("start on :8081")
	http.ListenAndServe(":8081", nil)
}
