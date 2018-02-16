package main

import (
	"container/list"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	"gitlab.com/rvasily/sphere-2017-2-hakaton2/exchange_broker_proto"
	"google.golang.org/grpc"
)

type (
	dataItem = struct {
		Ticker string
		Time   time.Time
		Price  float32
		Vol    int32
	}
	Request = struct {
		ClientID       int32
		Amount         int32
		Partial        bool
		Price          float32
		ResponseStream exchange_broker_proto.ExchangeBroker_GetDealServer
	}
	allRequests = struct {
		sync.Mutex
		m map[string]*list.List
	}
)

var (
	reqs            allRequests
	subscribersLock sync.RWMutex
	subscribers     []exchange_broker_proto.ExchangeBroker_OHLCVstreamServer
	agregatedOHLCV  map[string]exchange_broker_proto.OHLCV
	dataFiles       = []string{
		"data/SPFB.BR-1.18_171208_171208.txt",
		"data/SPFB.Si-12.17_171208_171208.txt",
		"data/SPFB.RTS-12.17_171208_171208.txt",
	}
)

func strToDataItem(s []string) dataItem {
	var res dataItem
	res.Ticker = s[0]
	t, err := time.Parse("150405", s[3])
	if err != nil {
		log.Panic(err)
	}
	n := time.Now()
	t = time.Date(n.Year(), n.Month(), n.Day(), t.Hour(), t.Minute(), t.Second(), 0, n.Location())
	res.Time = t
	price, err := strconv.ParseFloat(s[4], 64)
	if err != nil {
		log.Panic(err)
	}
	res.Price = float32(price)
	vol, err := strconv.Atoi(s[5])
	if err != nil {
		log.Panic(err)
	}
	res.Vol = int32(vol)
	return res
}

func DataStream() chan dataItem {
	var ch = make(chan dataItem, 1024)
	for _, fileName := range dataFiles {
		go func(fileName string) {
			f, err := os.Open(fileName)
			if err != nil {
				log.Panic(err)
			}
			reader := csv.NewReader(f)
			_, err = reader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Panic(err)
			}
			for {
				record, err := reader.Read()
				if err != nil {
					if err == io.EOF {
						return
					}
					log.Panic(err)
				}
				item := strToDataItem(record)
				diff := time.Since(item.Time)
				if diff < 0 {
					time.Sleep(-diff)
				}
				ch <- item
			}
		}(fileName)
	}
	return ch
}

type ExchangeServer struct{}

func (srv *ExchangeServer) OHLCVstream(nothing *exchange_broker_proto.Nothing, ostream exchange_broker_proto.ExchangeBroker_OHLCVstreamServer) error {
	log.Println("Subscriber added")
	subscribersLock.Lock()
	subscribers = append(subscribers, ostream)
	subscribersLock.Unlock()
	select {}
}

func (srv *ExchangeServer) GetDeal(iostream exchange_broker_proto.ExchangeBroker_GetDealServer) error {
	log.Println("Broker added")
	for {
		deal, err := iostream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Panic(err)
		}
		req := Request{
			deal.ClientID,
			deal.Amount,
			false,
			deal.Price,
			iostream,
		}
		reqs.m[deal.Ticker].PushBack(req)
	}
}

func agregateSyntheticDataAndSplitBuySell(inCh chan dataItem) chan dataItem {
	outCh := make(chan dataItem, 1024)
	go func() {
		prev := <-inCh
		for iitem := range inCh {
			switch {
			case iitem.Price == prev.Price && prev.Vol < 0:
				iitem.Vol *= -1
			case iitem.Price < prev.Price:
				iitem.Price *= -1
			}
			subscribersLock.Lock()
			defer subscribersLock.Unlock()
			if iitem.Price > agregatedOHLCV[iitem.Ticker].High {
				cur := agregatedOHLCV[iitem.Ticker]
				cur.High = iitem.Price
				agregatedOHLCV[iitem.Ticker] = cur
			}
			if iitem.Price < agregatedOHLCV[iitem.Ticker].Low {
				cur := agregatedOHLCV[iitem.Ticker]
				cur.Low = iitem.Price
				agregatedOHLCV[iitem.Ticker] = cur
			}
			cur := agregatedOHLCV[iitem.Ticker]
			cur.Volume += iitem.Vol
			cur.Close = iitem.Price
			agregatedOHLCV[iitem.Ticker] = cur
			prev = iitem
			outCh <- iitem
		}
	}()
	return outCh
}

var ID int64 = 0

func Agregate() {
	for {
		subscribersLock.Lock()
		log.Println("Sending stat to subscriber")
		for _, v := range agregatedOHLCV {
			v.Time = int32(time.Now().Hour()*10000 + time.Now().Minute()*100 + time.Now().Second())
			v.ID = ID
			ID = ID + 1
			v.Interval = 1
			for _, s := range subscribers {
				s.Send(&v)
			}
			v.Open = v.Close
			v.Volume = 0
			v.Low = v.Close
			v.High = v.Close
		}
		subscribersLock.Unlock()
		time.Sleep(time.Second)
	}
}

func DoTransactions(in chan dataItem, bids *allRequests) {
	for data := range in {
		log.Println("Diong transactions...")
		bids.Lock()
		req := bids.m[data.Ticker]
		for e := req.Front(); e != nil; {
			val, err := e.Value.(Request)
			if !err {
				log.Panic("Cant convert to Request: err")
			}
			if data.Vol*val.Amount < 0 {
				if (data.Vol > 0 && data.Price >= val.Price) || (data.Vol < 0 && data.Price <= val.Price) {
					vol := data.Vol + val.Amount
					if vol != 0 && data.Vol*vol < 0 {
						val.Amount = vol
						data.Vol = 0
					} else if vol != 0 && data.Vol*vol > 0 {
						val.Amount = 0
						data.Vol = vol
					} else {
						data.Vol = 0
						val.Amount = 0
					}
					data.Vol += val.Amount
					val.Amount += data.Vol
					if val.Amount != 0 {
						val.Partial = true
					}
					t := time.Now()
					strTime := t.Hour()*10000 + t.Minute()*100 + t.Second()
					out := &exchange_broker_proto.Deal{
						ClientID: val.ClientID,
						Ticker:   data.Ticker,
						Amount:   val.Amount,
						Price:    val.Price,
						Partial:  val.Partial,
						Time:     int32(strTime),
					}
					if data.Price != val.Price {
						out.Price = data.Price
					}
					val.ResponseStream.Send(out)
					if val.Amount == 0 {
						prev := e
						e = e.Next()
						req.Remove(prev)
					}
				}
			}
			if data.Vol == 0 {
				break
			}
		}
		bids.Unlock()
	}
}

func main() {
	agregatedOHLCV = make(map[string]exchange_broker_proto.OHLCV)
	reqs.m = make(map[string]*list.List)
	reqs.m["SPFB.BR-1.18"] = list.New()
	reqs.m["SPFB.RTS-12.17"] = list.New()
	reqs.m["SPFB.Si-12.17"] = list.New()
	go Agregate()
	go DoTransactions(agregateSyntheticDataAndSplitBuySell(DataStream()), &reqs)

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalln("cant listen port", err)
	}

	server := grpc.NewServer()

	st := &ExchangeServer{}

	exchange_broker_proto.RegisterExchangeBrokerServer(server, st)

	fmt.Println("starting server at :8081")
	server.Serve(lis)
}
