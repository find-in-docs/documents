package conn

import (
	"fmt"
	"os"

	"github.com/samirgadkari/sidecar/protos/v1/messages"
	"google.golang.org/grpc"
)

type SC struct{
	client *sidecarClient
}

func Connect() (*SC, error) {

	var opts []grpc.DialOption

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		fmt.Printf("Error creating GRPC channel\n:\terr: %v\n", err)
		os.Exit(-1)
	}

	defer conn.Close()

	client := messages.NewSidecarClient(conn)

	return &SC{client}, nil
}

func (sc *SC) Register() error {

	circuitConsecutiveFailures := 3

	// Go Duration is in the time package: https://pkg.go.dev/time#Duration
	// Go Duration maps to protobuf Duration.
	// You can convert between them using durationpb:
	//   https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb
	debounceDelayDuration, err := ParseDuration("5s") 
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	debounceDelay := durationpb.New(debounceDelayDuration)

	retryNum := 2

	retryDelayDuration, err := ParseDuration("2s")
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	retryDelay := durationpb.New(retryDelayDuration)

	rMsg := &messages.RegistrationMsg{
		Header.srcServType: 1,
		header.dstServType: 2,
		header.servId: 3,
		header.msgId: 4,

		CircuitFailureThrehold: &circuitConsecutiveFailures,
		DebounceDelay: &debounceDelay,
		RetryNum: &retryNum,
		RetryDelay: &retryDelay
	}

	sc.client.Register(context.Background(), rMsg)
}
