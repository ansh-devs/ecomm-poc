package server

import (
	"fmt"
	baserepo "github.com/ansh-devs/microservices_project/product-service/api/gapi"
	db "github.com/ansh-devs/microservices_project/product-service/db/generated"
	baseproto "github.com/ansh-devs/microservices_project/product-service/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

func MustStartGrpcServer(dbConn *db.Queries, grpcAddr string, loginsrvcport int) {

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	srvr := baserepo.NewGrpcServer(dbConn, loginsrvcport)
	grpcInterceptor := grpc.UnaryInterceptor(baserepo.GrpcInterceptor)
	grpcServer := grpc.NewServer(grpcInterceptor)
	baseproto.RegisterProductServiceServer(grpcServer, srvr)
	fmt.Printf("[INFO]: gRPC server started at : %s \n", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
