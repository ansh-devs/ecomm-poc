package server

import (
	"context"
	baserepo "github.com/ansh-devs/microservices_project/product-service/api/gapi"
	db "github.com/ansh-devs/microservices_project/product-service/db/generated"
	baseproto "github.com/ansh-devs/microservices_project/product-service/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
)

func MustStartHttpServer(dbConn *db.Queries, httpAddr string, loginsrvcport int) {
	muxOpts := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	gwmux := runtime.NewServeMux(muxOpts)
	srvr := baserepo.NewGrpcServer(dbConn, loginsrvcport)
	ctx, cancel := context.WithCancel(context.Background())
	err := baseproto.RegisterProductServiceHandlerServer(ctx, gwmux, srvr)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	httpConn, err := net.Listen("tcp", httpAddr)
	if err != nil {
		log.Fatalf("[ERROR]: can't listen on http address : %s", err)
	}
	httpMux := http.NewServeMux()
	handler := baserepo.HttpMiddleware(gwmux)
	httpMux.Handle("/", handler)
	log.Printf("[INFO]: gRPC-Gateway started at : %s \n", httpConn.Addr().String())
	if err := http.Serve(httpConn, httpMux); err != nil {
		log.Fatalf("[ERROR]: can't start http server : %s", err)
	}
	defer cancel()
}
