package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansh-devs/commercelens/login-service/db"
	"github.com/ansh-devs/commercelens/login-service/endpoints"
	"github.com/ansh-devs/commercelens/login-service/repo"
	"github.com/ansh-devs/commercelens/login-service/service"
	"github.com/ansh-devs/commercelens/login-service/transport"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	_ "github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	var httpAddr = os.Getenv("HTTPPORT")
	tracer := opentracing.GlobalTracer()
	cfg := &config.Configuration{
		ServiceName: "login-service",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}

	errs := make(chan error)
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		errs <- fmt.Errorf("%s", err)
	}
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			errs <- fmt.Errorf("%s", err)
		}
	}(closer)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "login_service",
			"time", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}
	_ = level.Info(logger).Log("msg", "service started")

	ctx := context.Background()
	var srv *service.LoginService
	{
		var dbSource = fmt.Sprintf("postgres://%s:%s@%s/%s",
			os.Getenv("DBUSER"),
			os.Getenv("DBPASSWORD"),
			os.Getenv("DBHOST"),
			os.Getenv("DBNAME"),
		)
		dbConn := db.MustConnectToPostgress(dbSource)
		repository := repo.NewPostgresRepo(dbConn, logger, tracer)
		srv = service.NewService(repository, logger, tracer)
	}
	go srv.GetUserWithNats(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	endpoint := endpoints.NewEndpoints(srv)
	srv.RegisterService(&httpAddr)
	go srv.UpdateHealthStatus()

	go func() {
		fmt.Println("listening on port", httpAddr)
		handler := transport.NewHttpServer(ctx, endpoint)
		errs <- http.ListenAndServe(httpAddr, handler)
	}()

	for sig := range errs {
		_ = level.Error(logger).Log("status", sig, "GRACEFULLY SHUTTING DOWN !")
		_ = srv.ConsulClient.Agent().ServiceDeregister(srv.SrvID)
		os.Exit(0)
	}

}
