package grcp_client

import (
	"context"
	"fmt"
	"io"
	"sphere-2017-2-hakaton2/database"
	"sphere-2017-2-hakaton2/exchange_broker_proto"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type Nothing struct {
	Dummy bool
}

func GrpcClient(ch chan database.Request) {

	grcpConn, err := grpc.Dial(
		"127.0.0.1:8081",
		grpc.WithInsecure(),
	)
	if err != nil {
		fmt.Println("can't connect to server")
	}
	defer grcpConn.Close()

	cl1 := exchange_broker_proto.NewExchangeBrokerClient(grcpConn)

	ctx1 := context.Background()
	client1, err := cl1.GetDeal(ctx1)
	if err != nil {
		fmt.Println("some problem")
	}
	obj := exchange_broker_proto.Nothing{
		Dummy: false,
	}
	ctx2 := context.Background()
	client2, err := cl1.OHLCVstream(ctx2, &obj)
	if err != nil {
		fmt.Println("some problem")
	}

	wg := &sync.WaitGroup{}
	wg.Add(3)

	// отсылаем новую заявку
	go func(wg *sync.WaitGroup) {
		deal := exchange_broker_proto.Deal{}
		for req := range ch {
			deal.ClientID = int32(req.UserID)
			deal.Ticker = req.Ticker
			deal.Amount = int32(req.Vol)
			deal.Time = int32(time.Now().Unix())
			deal.Price = req.Price
			client1.Send(&deal)
			fmt.Println("отправил заявку от", req.UserID)
		}
	}(wg)

	// принимаем изменение заявки
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			recv, err := client1.Recv()
			if err == io.EOF {
				fmt.Println("\tstream closed")
				return
			} else if err != nil {
				fmt.Println("\terror happend", err)
				return
			}
			fmt.Println(" <-", recv)
			// обратиться к базе
			client, err := database.GetClient(int(recv.ClientID))
			client.Balance += int(recv.Price) * int(recv.Amount)
			database.UpdateClient(client)
			positions := database.GetPositions(client.ID)
			for i := range positions {
				if positions[i].Ticker == recv.Ticker {
					positions[i].Vol += int(recv.Amount)
				}
			}
			database.UpdatePositions(int(recv.ClientID), positions)
			transaction := database.Order{}
			transaction.Time = int(recv.Time)
			transaction.UserID = int(recv.ClientID)
			transaction.Ticker = recv.Ticker
			if recv.Amount > 0 {
				transaction.IsBuy = 1
				transaction.Vol = int(recv.Amount)
			} else {
				transaction.IsBuy = 0
				transaction.Vol = int(recv.Amount * -1.0)
			}
			transaction.Price = recv.Price
			database.InsertOrder(transaction)
		}
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			recvInfo, err := client2.Recv()
			if err == io.EOF {
				fmt.Println("\tstream closed")
				return
			} else if err != nil {
				fmt.Println("\terror happend", err)
				return
			}
			fmt.Println(" <-", recvInfo)
			err = database.InsertOLHCV(*recvInfo)
			if err != nil {
				panic(err)
			}
			fmt.Println("-> ", recvInfo)
		}
	}(wg)
	wg.Wait()
}
