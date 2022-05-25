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

	topic := "search.doc.import.v1"
	workQueue := "uploadWorkQueue"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = sidecar.SubJS(ctx, topic, workQueue, allTopicsRecvChanSize)
	if err != nil {
		fmt.Printf("Error subscribing to topic: %s err: %v\n", topic, err)
		os.Exit(-1)
	}

	time.Sleep(time.Second)
	recvCh := sidecar.RecvJS(ctx, topic, workQueue)
	count := 0
	for {
		select {
		case m, ok := <-recvCh:
			if !ok {
				fmt.Printf("Error receiving from channel - cancelling context\n")
				cancel()
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
