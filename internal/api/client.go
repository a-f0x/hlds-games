package api

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetServerInfo(host string, port int64) func(ctx context.Context) (*GameStatus, error) {
	url := fmt.Sprintf("%s:%d", host, port)
	return func(ctx context.Context) (*GameStatus, error) {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		client := NewHalfLifeDedicatedServerClient(conn)
		status, responseError := client.GetServerStatus(ctx, &GetGameStatusRequest{})
		if responseError != nil {
			return nil, responseError
		}
		return status, nil
	}
}

func ExecuteRconCommand(host string, port int64) func(ctx context.Context, command string) (*RconCommandResult, error) {
	url := fmt.Sprintf("%s:%d", host, port)
	return func(ctx context.Context, command string) (*RconCommandResult, error) {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		client := NewHalfLifeDedicatedServerClient(conn)
		status, responseError := client.ExecuteRconCommand(ctx, &RconCommand{Command: command})
		if responseError != nil {
			return nil, responseError
		}
		return status, nil
	}
}
