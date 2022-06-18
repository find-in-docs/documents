package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/find-in-docs/documents/pkg/config"
	"github.com/find-in-docs/documents/pkg/data"
	"github.com/find-in-docs/documents/pkg/transform"
	"github.com/find-in-docs/sidecar/pkg/client"
	"github.com/find-in-docs/sidecar/pkg/utils"
	pb "github.com/find-in-docs/sidecar/protos/v1/messages"
	"github.com/spf13/viper"
)

const (
	allTopicsRecvChanSize = 32
	maxMsgLen             = 130
)

func copyDoc(d *data.Doc, s *pb.Doc) {

	d.DocId = data.DocumentId(s.DocId)
	d.WordInts = make([]data.WordInt, len(s.WordInts))
	for i, v := range s.WordInts {
		d.WordInts[i] = data.WordInt(v)
	}
	d.InputDocId = s.InputDocId
	d.UserId = s.UserId
	d.BusinessId = s.BusinessId
	d.Stars = s.Stars
	d.Useful = uint16(s.Useful)
	d.Funny = uint16(s.Funny)
	d.Cool = uint16(s.Cool)
	d.Text = s.Text
	d.Date = s.Date
}

func main() {

	var wordInts []data.WordInt
	var wordToInt map[string]data.WordInt

	config.Load()

	stopwords := data.LoadStopwords(viper.GetString("englishStopwordsFile"))
	wordsToInts := transform.WordsToInts(stopwords)

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

	processDocs := make(chan *data.Doc)

	go func() {
	LOOP:
		for {
			select {
			case docs, ok := <-recvDocs:
				if !ok {
					fmt.Printf("Receive docs channel closed:\n\t"+
						"Reason: %s\n", ctx.Err())
					os.Exit(-1)
				}

				fmt.Printf(".")

				for _, doc := range docs.Documents.Doc {
					wordInts, wordToInt = wordsToInts(doc.Text)

					var d data.Doc
					copyDoc(&d, doc)
					d.WordInts = wordInts
					processDocs <- &d

					if err := db.StoreData(&d, tableName, wordInts); err != nil {
						break
					}
				}

			case <-ctx.Done():

				if err := db.DBDisconnect(); err != nil {
					break LOOP
				}

				break LOOP
			}
		}
	}()

	if err := db.StoreWordIntMappings("word_to_int", wordToInt); err != nil {
		fmt.Printf("Error storing word-to-int mappings: %v", err)
		os.Exit(-1)
	}

	fmt.Printf("Transforming WordToDocs\n")
	if err := transform.WordToDocs(processDocs, db.StoreWordToDocMappings); err != nil {
		fmt.Printf("Error creating word-to-doc mapping: %v", err)
		os.Exit(-1)
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
