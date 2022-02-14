package conn

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/samirgadkari/sidecar/protos/v1/messages"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

type SC struct {
	client messages.SidecarClient
}

func Connect(serverAddr string) (*SC, error) {

	// var opts []grpc.DialOption

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error creating GRPC channel\n:\terr: %v\n", err)
		os.Exit(-1)
	}

	defer conn.Close()

	client := messages.NewSidecarClient(conn)

	return &SC{client}, nil
}

func (sc *SC) Register() error {

	var circuitConsecutiveFailures uint32 = 3

	// Go Duration is in the time package: https://pkg.go.dev/time#Duration
	// Go Duration maps to protobuf Duration.
	// You can convert between them using durationpb:
	//   https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb
	debounceDelayDuration, err := time.ParseDuration("5s")
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	debounceDelay := durationpb.New(debounceDelayDuration)

	var retryNum uint32 = 2

	retryDelayDuration, err := time.ParseDuration("2s")
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	retryDelay := durationpb.New(retryDelayDuration)
	header := messages.Header{
		SrcServType: "postgresService",
		DstServType: "sidecarService",
		ServId:      0,
		MsgId:       0,
	}

	rMsg := &messages.RegistrationMsg{
		Header: &header,

		CircuitFailureThreshold: &circuitConsecutiveFailures,
		DebounceDelay:           debounceDelay,
		RetryNum:                &retryNum,
		RetryDelay:              retryDelay,
	}

	sc.client.Register(context.Background(), rMsg)

	return nil
}
