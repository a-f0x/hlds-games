package api

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ExecuteRconCommand(address string) func(ctx context.Context, command string) (*RconCommandResult, error) {
	return func(ctx context.Context, command string) (*RconCommandResult, error) {
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
