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
	
	rMsg := new(RegistrationMsg)
	rMsg.Header.srcServType = 1
	rMsg.header.dstServType = 2
	rMsg.header.servId = 3
	rMsg.header.msgId = 4

	circuit := Circuit{ failureThreshold: 5 }
	debounce := Debounce{ delay: 2 }
	retry := Retry{ Retries: 3, delay: 1 }

	rMsg.limits = [][circuit, debounce, retry]

	sc.client.Register(context.Background(), rMsg)
}
