package database

import (
	"database/sql"

	"sphere-2017-2-hakaton2/exchange_broker_proto"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// DSN это соединение с базой
	// вы можете изменить этот параметр, в тестах соединение будет браться отсюда
	DSN = "root@(localhost:3306)/hackaton?charset=utf8&interpolateParams=true"
)

type Client struct {
	ID      int
	LoginID int
	Balance int
}

type Position struct {
	ID     int    `json:"id"`
	UserID int    `json:"-"`
	Ticker string `json:"ticker"`
	Vol    int    `json:"vol"`
}

type Order struct {
	ID     int     `json:"id"`
	Time   int     `json:"-"`
	UserID int     `json:"-"`
	Ticker string  `json:"ticker"`
	Vol    int     `json:"vol"`
	IsBuy  int     `json:"is_buy"`
	Price  float32 `json:"price"`
}

type Request struct {
	ID     int     `json:"id"`
	UserID int     `json:"u_id"`
	Ticker string  `json:"ticker"`
	Vol    int     `json:"vol"`
	Price  float32 `json:"price"`
	IsBuy  int     `json:"is_buy"`
}

type Stat struct {
	ID       int
	Time     int
	Interval int
	Open     float32
	High     float32
	Low      float32
	Close    float32
	Volume   int
	Ticker   string
}

func InitDB() *sql.DB {
	db, err := sql.Open("mysql", DSN)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}
	return db
}

func CreateClient(loginID, cash int) (int, error) {
	db := InitDB()
	var err error

	client, err := GetClient(loginID)
	if err != nil {
		result, err := db.Exec("INSERT INTO clients (login_id, balance) values (?, ?)", loginID, cash)
		if err != nil {
			return 0, nil
		}
		id, err := result.LastInsertId()
		return int(id), err
	}
	return client.ID, nil
}

func GetClient(ID int) (Client, error) {
	db := InitDB()

	result := db.QueryRow("SELECT * from clients where id = ? order by desc", ID)

	client := Client{}
	err := result.Scan(&client.ID, &client.LoginID, &client.Balance)
	if err != nil {
		return client, err
	}
	return client, nil
}

func GetClientByLoginID(loginID int) (Client, error) {
	db := InitDB()

	result := db.QueryRow("SELECT * from clients where login_id = ?", loginID)

	client := Client{}
	err := result.Scan(&client.ID, &client.LoginID, &client.Balance)
	if err != nil {
		return client, err
	}
	return client, nil
}

func UpdateClient(client Client) error {
	db := InitDB()

	_, err := db.Exec("UPDATE clients set balance = ? where id = ?", client.Balance, client.ID)
	if err != nil {
		panic(err)
	}
	return nil
}

func GetPositions(clientID int) []Position {
	db := InitDB()
	rows, err := db.Query("select * FROM positions where user_id = ?", clientID)

	if err != nil {
		panic(err)
	}

	posititions := []Position{}

	for rows.Next() {
		newPosition := Position{}
		rows.Scan(&newPosition.ID, &newPosition.UserID, &newPosition.Ticker, &newPosition.Vol)
		posititions = append(posititions, newPosition)
	}

	return posititions
}

func UpdatePositions(ClientID int, positions []Position) error {
	db := InitDB()

	for _, pos := range positions {
		result := db.QueryRow("SELECT * from positions where id = ?", pos.ID)
		var tmp interface{}
		err := result.Scan(&tmp, &tmp, &tmp, &tmp)
		if err == sql.ErrNoRows {
			_, err2 := db.Exec("INSERT INTO positions (user_id, ticker, vol) values (?, ?, ?)", pos.UserID, pos.Ticker, pos.Vol)
			if err2 != nil {
				panic(err2)
			}
		} else if err != nil {
			return err
		} else {
			_ = db.QueryRow("UPDATE positions set volume = ? where id = ?")
		}

	}

	return nil
}

func GetOrderByID(clientID int) []Order {
	db := InitDB()
	rows, err := db.Query("select * FROM orders_history where user_id = ? order by time", clientID)

	if err != nil {
		panic(err)
	}

	orders := []Order{}

	for rows.Next() {
		newOrder := Order{}
		rows.Scan(&newOrder.ID, &newOrder.Time, &newOrder.UserID, &newOrder.Ticker, &newOrder.Vol, &newOrder.Price, &newOrder.IsBuy)
		orders = append(orders, newOrder)
	}

	return orders
}

func GetOrders() []Order {
	db := InitDB()
	rows, err := db.Query("select * FROM orders_history order by time")

	if err != nil {
		panic(err)
	}

	orders := []Order{}

	for rows.Next() {
		newOrder := Order{}
		rows.Scan(&newOrder.ID, &newOrder.Time, &newOrder.UserID, &newOrder.Ticker, &newOrder.Vol, &newOrder.Price, &newOrder.IsBuy)
		orders = append(orders, newOrder)
	}

	return orders
}

func InsertOrder(order Order) error {
	db := InitDB()
	_, err := db.Exec("INSERT INTO orders_history (time, user_id, ticker, vol, price, is_buy) values (?, ?, ?, ?, ?, ?)", order.Time, order.UserID, order.Ticker, order.Vol, order.Price, order.IsBuy)

	if err != nil {
		return err
	}
	return nil
}

func GetRequests(clientID int) []Request {
	db := InitDB()
	rows, err := db.Query("select * FROM request where user_id = ?", clientID)

	if err != nil {
		panic(err)
	}

	requests := []Request{}

	for rows.Next() {
		newRequest := Request{}
		rows.Scan(&newRequest.ID, &newRequest.UserID, &newRequest.Ticker, &newRequest.Vol, &newRequest.Price, &newRequest.IsBuy)
		requests = append(requests, newRequest)
	}

	return requests
}

func InsertRequest(request Request) (int, error) {
	db := InitDB()
	result, err := db.Exec("INSERT INTO request (user_id, ticker, vol, price, is_buy) values (?, ?, ?, ?, ?)", request.UserID, request.Ticker, request.Vol, request.Price, request.IsBuy)

	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func DeleteRequest(UserID, ReqID int) error {
	db := InitDB()
	_, err := db.Exec("DELETE from request where id = ?", ReqID)

	if err != nil {
		return err
	}
	return nil
}

func InsertOLHCV(stat exchange_broker_proto.OHLCV) error {
	db := InitDB()
	_, err := db.Exec("INSERT INTO stat values (?, ?, ?, ?, ?, ?, ?, ?)",
		stat.Interval, stat.Open, stat.High, stat.Low, stat.Close, stat.Volume, stat.Ticker)

	if err != nil {
		return err
	}

	return nil
}

func GetStat() []Stat {
	db := InitDB()
	rows, err := db.Query("SELECT * from stat order by id desc limit 300")

	if err != nil {
		panic(err)
	}

	stats := []Stat{}

	for rows.Next() {
		newStat := Stat{}
		rows.Scan(&newStat.ID, &newStat.Time, &newStat.Interval, &newStat.Open, &newStat.High, &newStat.Low, &newStat.Close, &newStat.Volume, &newStat.Ticker)
		stats = append(stats, newStat)
	}

	return stats
}

// func main() {
// 	id, err := CreateClient(123123, 123)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(id)

// 	client, err := GetClient(id)
// 	otherClient, err := GetClient(333333)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("%#v\n", otherClient)

// 	client.Balance = 0
// 	err = UpdateClient(client)
// 	if err != nil {
// 		panic(err)
// 	}

// 	client, err = GetClient(id)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("%#v\n", client)

// 	positions := []Position{}
// 	positions = append(positions, Position{
// 		UserID: client.ID,
// 		Ticker: "oil",
// 		Vol:    123,
// 	})

// 	err = UpdatePositions(client.ID, positions)
// 	if err != nil {
// 		panic(err)
// 	}

// 	gotPositions := GetPositions(client.ID)
// 	fmt.Printf("%#v\n", gotPositions)

// 	positions = append(positions, Position{
// 		UserID: client.ID,
// 		Ticker: "oil",
// 		Vol:    123,
// 	})

// 	err = InsertOrder(Order{
// 		Time:   123123,
// 		UserID: 123123,
// 		Ticker: "oil",
// 		Vol:    123,
// 		Price:  123,
// 		IsBuy:  1,
// 	})

// 	if err != nil {
// 		panic(err)
// 	}

// 	gotOrders := GetOrders()
// 	gotOrders2 := GetOrderByID(123123)
// 	fmt.Printf("%#v\n", gotOrders)
// 	fmt.Printf("%#v\n", gotOrders2)

// 	err = InsertRequest(Request{
// 		UserID: 12312123,
// 		Ticker: "oil",
// 		Vol:    32131231,
// 		Price:  123123,
// 		IsBuy:  0,
// 	})

// 	gotRequests := GetRequests(12312123)
// 	gotRequests2 := GetRequests(123121233123)
// 	fmt.Printf("%#v\n", gotRequests)
// 	fmt.Printf("%#v\n", gotRequests2)
// }
