package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tm "github.com/buger/goterm"
	pb "github.com/justanotherorganization/tiberious/proto/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	grpcAddr = flag.String("grpc_address", "localhost:4004", "the grpc address")
)

func readStreamMessage(stream pb.Tiberious_StartStreamClient) {
	defer tm.Flush()
	for {
		in, err := stream.Recv()
		if err != nil {
			stat, ok := status.FromError(err)
			if ok && stat.Code() == codes.Unavailable {
				tm.Print("stream closed by server")
				return
			}

			tm.Println(err.Error())
			return
		}

		if msg := in.GetClientMessage(); msg != nil {
			tm.Printf("%v", msg)
			tm.Flush()
		}
	}
}

func sendMessage(conversationID int64, msg string, stream pb.Tiberious_StartStreamClient) error {
	m := pb.ClientMessage{
		Meta: &pb.Meta{
			Time: pb.NewTimestamp(time.Now()),
		},
		ConversationId: conversationID,
		Data:           []byte(msg),
	}

	return stream.Send(&pb.StreamMessage{
		StreamMessage: &pb.StreamMessage_ClientMessage{
			ClientMessage: &m,
		},
	})
}

func handleCommand(cmd string, stream pb.Tiberious_StartStreamClient) error {
	switch {
	// Keep in alpha
	case strings.HasPrefix(cmd, "msg"):
		cmd = strings.TrimSpace(strings.TrimPrefix(cmd, "msg"))
		slice := strings.SplitN(cmd, " ", 2)
		if l := len(slice); l != 2 {
			return errors.Errorf("len(slice) %d should be equal to 2", l)
		}

		cmd = slice[1]
		id, err := strconv.Atoi(slice[0])
		if err != nil {
			return errors.Wrap(err, "strconv.Atoi")
		}

		return sendMessage(int64(id), cmd, stream)
	default:
		tm.Println("Unknown command:", cmd)
	}

	tm.Flush()
	return nil
}

func handleInput(input *bufio.Reader, stream pb.Tiberious_StartStreamClient) error {
	in, _ := input.ReadString('\n')
	switch {
	case strings.HasPrefix(in, "/"):
		return handleCommand(strings.TrimPrefix(in, "/"), stream)
	default:
		break
	}
	return nil
}

func main() {
	tm.Clear()
	tm.Flush()
	flag.Parse()

	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	client := pb.NewTiberiousClient(conn)

	stream, err := client.StartStream(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}
	go readStreamMessage(stream)

	input := bufio.NewReader(os.Stdin)

	var running = true
	for running {
		if err = handleInput(input, stream); err != nil {
			fmt.Println(err.Error())
			running = false
		}
	}
}
