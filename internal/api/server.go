package api

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"hlds-games/internal/config"
	"hlds-games/internal/rcon"
	"log"
	"net"
)

type HLDSApiServer struct {
	gameType      string
	grpcApiConfig *config.GrpcApiConfig
	rcon          *rcon.Rcon
}

func NewHLDSApiServer(gameType string, grpcApiConfig *config.GrpcApiConfig, rcon *rcon.Rcon) *HLDSApiServer {
	server := &HLDSApiServer{
		gameType:      gameType,
		grpcApiConfig: grpcApiConfig,
		rcon:          rcon,
	}
	return server
}

func (h *HLDSApiServer) RunServer() {
	url := fmt.Sprintf(":%d", h.grpcApiConfig.GrpcApiPort)
	lis, err := net.Listen("tcp", url)
	if err != nil {
		log.Fatalf("fail to listen port: %d, %v", h.grpcApiConfig.GrpcApiPort, err)
	}
	grpcServer := grpc.NewServer()
	RegisterHalfLifeDedicatedServerServer(grpcServer, h)
	log.Println(fmt.Sprintf("Server started %s", url))
	grpcServerError := grpcServer.Serve(lis)

	if grpcServerError != nil {
		log.Fatalf("fail to start server %v", grpcServerError)
	}

}

func (h *HLDSApiServer) GetServerStatus(context.Context, *GetGameStatusRequest) (*GameStatus, error) {
	result, rconError := h.rcon.GetServerStatus()
	if rconError != nil {
		return nil, rconError
	}
	return &GameStatus{
		Game:    h.gameType,
		Name:    result.Name,
		Map:     result.Map,
		Players: result.Players,
	}, nil
}

func (h *HLDSApiServer) ExecuteRconCommand(ctx context.Context, request *RconCommand) (*RconCommandResult, error) {
	result, rconError := h.rcon.SendRconCommand(request.Command)
	if rconError != nil {
		return nil, rconError
	}
	return &RconCommandResult{
		Result: *result,
	}, nil
}

func GetInfo(host string, port int64) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithBlock(),
	)
	if err != nil {
		log.Println(err)
		log.Fatalf(fmt.Sprintf("connection to server failed. %s", err))
	}
	defer conn.Close()
	client := NewHalfLifeDedicatedServerClient(conn)
	status, responseError := client.GetServerStatus(context.Background(), &GetGameStatusRequest{})
	if responseError != nil {
		log.Fatalf("Response error %s", responseError)
	}
	log.Printf("response: %v", status)
}
