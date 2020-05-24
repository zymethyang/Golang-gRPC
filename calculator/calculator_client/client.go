package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"../calculatorpb"
	"google.golang.org/grpc"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()
	c := calculatorpb.NewCalculateServiceClient(cc)

	doBiDiStreaming(c)
}

func doUnary(c calculatorpb.CalculateServiceClient) {
	fmt.Println("Starting an unary request")

	req := &calculatorpb.SumRequest{
		Sum: &calculatorpb.Sum{
			FirstNumber:  3,
			SecondNumber: 7,
		},
	}

	res, err := c.Sum(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling Sum RPC: %v", err)
	}

	log.Printf("Response from Sum: %v", res.Result)
}

func doServerStreaming(c calculatorpb.CalculateServiceClient) {
	fmt.Println("Starting an streaming request")

	req := &calculatorpb.PrimeNumberDecRequest{
		PrimeNumberDec: &calculatorpb.PrimeNumberDec{
			Number: 120,
		},
	}

	resStream, err := c.PrimeNumberDec(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling PrimeNumberDec RPC: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while read stream: %v", err)
		}
		log.Printf("Response from PrimeNumberDec: %v", msg.GetResult())
	}
}

func doClientStreaming(c calculatorpb.CalculateServiceClient) {
	fmt.Printf("Starting to do a Client Streaming RPC... \n")

	requests := []*calculatorpb.ComputeAverageRequest{
		&calculatorpb.ComputeAverageRequest{
			ComputeAverage: &calculatorpb.ComputeAverage{
				Number: 1,
			},
		},
		&calculatorpb.ComputeAverageRequest{
			ComputeAverage: &calculatorpb.ComputeAverage{
				Number: 2,
			},
		},
		&calculatorpb.ComputeAverageRequest{
			ComputeAverage: &calculatorpb.ComputeAverage{
				Number: 3,
			},
		},
		&calculatorpb.ComputeAverageRequest{
			ComputeAverage: &calculatorpb.ComputeAverage{
				Number: 4,
			},
		},
	}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling ComputeAverage: %v \n", err)
	}

	for _, req := range requests {
		fmt.Printf("Sending req: %v \n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response from ComputeAverage: %v \n", err)
	}
	fmt.Printf("ComputeAverage Response: %v \n", res)
}

func doBiDiStreaming(c calculatorpb.CalculateServiceClient) {
	fmt.Printf("Starting to do a Streaming RPC... \n")

	// We create a stream by involking the client
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: ̀%v \n", err)
		return
	}

	requests := []*calculatorpb.FindMaximumRequest{
		&calculatorpb.FindMaximumRequest{
			FindMaximum: &calculatorpb.FindMaximum{
				Number: 1,
			},
		},
		&calculatorpb.FindMaximumRequest{
			FindMaximum: &calculatorpb.FindMaximum{
				Number: 5,
			},
		},
		&calculatorpb.FindMaximumRequest{
			FindMaximum: &calculatorpb.FindMaximum{
				Number: 3,
			},
		},
		&calculatorpb.FindMaximumRequest{
			FindMaximum: &calculatorpb.FindMaximum{
				Number: 6,
			},
		},
		&calculatorpb.FindMaximumRequest{
			FindMaximum: &calculatorpb.FindMaximum{
				Number: 2,
			},
		},
		&calculatorpb.FindMaximumRequest{
			FindMaximum: &calculatorpb.FindMaximum{
				Number: 20,
			},
		},
	}

	waitc := make(chan struct{})

	// We receive a bunch of message from the client (go routine)
	go func() {
		for _, req := range requests {
			fmt.Printf("Sending message: %v \n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				close(waitc)
			}
			fmt.Printf("Received: %v \n", res.GetResult())
		}
	}()

	// Block until everything is done
	<-waitc
}
