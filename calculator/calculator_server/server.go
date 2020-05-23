package main

import (
	"context"
	"log"
	"net"
	"time"

	"../calculatorpb"

	"google.golang.org/grpc"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	firstNumber := req.GetSum().GetFirstNumber()
	secondNumber := req.GetSum().GetSecondNumber()
	result := firstNumber + secondNumber
	res := &calculatorpb.SumResponse{
		Result: result,
	}
	return res, nil
}

func (*server) PrimeNumberDec(req *calculatorpb.PrimeNumberDecRequest, stream calculatorpb.CalculateService_PrimeNumberDecServer) error {
	number := req.GetPrimeNumberDec().GetNumber()
	k := (int32)(2)

	for number > 1 {
		if number%k == 0 {
			result := k
			res := &calculatorpb.PrimeNumberDecResponse{
				Result: result,
			}
			stream.Send(res)
			number = number / k
		} else {
			k = k + 1
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatal("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculateServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
