package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"../calculatorpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (*server) ComputeAverage(stream calculatorpb.CalculateService_ComputeAverageServer) error {
	fmt.Printf("ComputeAverage function was invoked with a streaming request \n")
	sum := (float32)(0)
	length := (float32)(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("Sum: %v", sum)
			fmt.Printf("Length: %v: ", length)
			// we have finished reading the client stream
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Result: (float32)(sum / length),
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v \n", err)
		}

		number := req.GetComputeAverage().GetNumber()
		fmt.Printf("Received number : %v \n", number)
		sum += number
		length++
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculateService_FindMaximumServer) error {
	fmt.Printf("FindMaximum function was invoked with a streaming request \n")
	max := (int32)(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		number := req.GetFindMaximum().GetNumber()
		if number > max {
			max = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Result: max,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending data to client stream: %v", sendErr)
				return sendErr
			}
		}
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("Received SquareRoot RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
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
