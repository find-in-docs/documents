package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/find-in-docs/documents/pkg/config"
	"github.com/find-in-docs/documents/pkg/data"
	"github.com/find-in-docs/sidecar/pkg/client"
	"github.com/find-in-docs/sidecar/pkg/utils"
	"github.com/spf13/viper"
)

const (
	allTopicsRecvChanSize = 32
	maxMsgLen             = 130
)

func main() {

	config.Load()

	// Setup database
	db, err := data.DBConnect()
	if err != nil {
		return
	}

	tableName := "documents"
	err = db.CreateDocumentsTable()
	if err != nil {
		return
	}

	sidecar, err := client.InitSidecar(tableName, nil)
	if err != nil {
		fmt.Printf("Error initializing sidecar: %v\n", err)
		os.Exit(-1)
	}

	subject := viper.GetString("nats.jetstream.subject")
	durableName := viper.GetString("nats.jetstream.consumer.durableName")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recvDocs, err := sidecar.ReceiveDocs(ctx, subject, durableName)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	for {
		select {
		case doc, ok := <-recvDocs:
			if !ok {
				fmt.Printf("Receive docs channel closed:\n\t"+
					"Reason: %s\n", ctx.Err())
				os.Exit(-1)
			}

			fmt.Printf("-- doc.MsgNumber: %d --", doc.MsgNumber)

		case <-ctx.Done():
			break
		}
	}

	/*
		err = sidecar.AddJS(ctx, subject, durableName)
		if err != nil {
			fmt.Printf("Error subscribing to subject: %s err: %v\n", subject, err)
			os.Exit(-1)
		}

		time.Sleep(time.Second)
		recvCh := sidecar.RecvJS(context.Background(), subject, durableName)
		count := 0
		for {
			select {
			case m, ok := <-recvCh:
				if !ok {
					fmt.Printf("Error receiving from channel\n")
					break
				}
				shortMsg := string(m.Response.Msg)
				l := maxMsgLen
				if len(shortMsg) < maxMsgLen {
					l = len(shortMsg)
				}
				fmt.Printf("Received message: %s\n", shortMsg[:l])
				count++
			case <-ctx.Done():
				break
			}
		}

		fmt.Printf(" >>>>>>> Received %d messages\n", count)
	*/

	/* This is an example of how to publish a message. It is a log message because for now
	* it is the only type that is received (by this same persistLogs service).

	sidecar.Logger.Log("Persist sending log message test: %s\n", "search.log.v1")
	time.Sleep(3 * time.Second)

	var retryNum uint32 = 1
	retryDelayDuration, err := time.ParseDuration("200ms")
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	retryDelay := durationpb.New(retryDelayDuration)

	err = sidecar.Pub(ctx, "search.data.v1", []byte("test pub message"),
		&pb.RetryBehavior{
			RetryNum:   &retryNum,
			RetryDelay: retryDelay,
		},
	)
	if err != nil {
		fmt.Printf("Error publishing message.\n\terr: %v\n", err)
	}

	*/

	fmt.Println("Press the Enter key to stop")
	fmt.Scanln()
	fmt.Println("User pressed Enter key")

	// Signal that we want the process subscription goroutines to end.
	// This cancellation causes the goroutines to unsubscribe from the topic
	// before they end themselves.
	// cancel()

	sleepDur, _ := time.ParseDuration("3s")
	fmt.Printf("Sleeping for %s seconds\n", sleepDur)
	time.Sleep(sleepDur)

	utils.ListGoroutinesRunning()

	select {} // This will wait forever
}
