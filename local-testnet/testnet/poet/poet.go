package poet

import (
	"fmt"
	proxy "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spacemeshos/poet/rpc"
	"github.com/spacemeshos/poet/rpc/api"
	"github.com/spacemeshos/poet/rpccore"
	"github.com/spacemeshos/poet/rpccore/apicore"
	"github.com/spacemeshos/poet/service"
	"github.com/spacemeshos/poet/signal"
	"github.com/spacemeshos/smutil/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	"net"
	"net/http"
	"os"
	"time"
)

// startServer starts the RPC server.
func StartServer(cfg *config, stopPoet chan struct{}) error {
	ce := make(chan error)
	go func() {
		sig := signal.NewSignal()
		ce <- func() error{
			// Resolve the RPC listener
			addr, err := net.ResolveTCPAddr("tcp", cfg.RawRPCListener)
			if err != nil {
				return err
			}
			cfg.RPCListener = addr

			// Resolve the REST listener
			addr, err = net.ResolveTCPAddr("tcp", cfg.RawRESTListener)
			if err != nil {
				return err
			}
			cfg.RESTListener = addr

			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			// Initialize and register the implementation of gRPC interface
			var grpcServer *grpc.Server
			var proxyRegstr []func(context.Context, *proxy.ServeMux, string, []grpc.DialOption) error
			options := []grpc.ServerOption{
				grpc.UnaryInterceptor(loggerInterceptor()),
				// XXX: this is done to prevent routers from cleaning up our connections (e.g aws load balances..)
				// TODO: these parameters work for now but we might need to revisit or add them as configuration
				// TODO: Configure maxconns, maxconcurrentcons ..
				grpc.KeepaliveParams(keepalive.ServerParameters{
					MaxConnectionIdle:     time.Minute * 120,
					MaxConnectionAge:      time.Minute * 180,
					MaxConnectionAgeGrace: time.Minute * 10,
					Time:                  time.Minute,
					Timeout:               time.Minute * 3,
				}),
			}

			if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
				if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
					return err
				}
			}

			if cfg.CoreServiceMode {
				rpcServer := rpccore.NewRPCServer(sig, cfg.DataDir)
				grpcServer = grpc.NewServer(options...)

				apicore.RegisterPoetCoreProverServer(grpcServer, rpcServer)
				apicore.RegisterPoetVerifierServer(grpcServer, rpcServer)
				proxyRegstr = append(proxyRegstr, apicore.RegisterPoetCoreProverHandlerFromEndpoint)
				proxyRegstr = append(proxyRegstr, apicore.RegisterPoetVerifierHandlerFromEndpoint)
			} else {
				svc, err := service.NewService(sig, cfg.Service, cfg.DataDir)
				if err != nil {
					return err
				}

				rpcServer := rpc.NewRPCServer(svc)
				grpcServer = grpc.NewServer(options...)

				api.RegisterPoetServer(grpcServer, rpcServer)
				proxyRegstr = append(proxyRegstr, api.RegisterPoetHandlerFromEndpoint)
			}

			// Start the gRPC server listening for HTTP/2 connections.
			lis, err := net.Listen(cfg.RPCListener.Network(), cfg.RPCListener.String())
			if err != nil {
				return fmt.Errorf("failed to listen: %v", err)
			}
			defer lis.Close()

			go func() {
				log.Info("RPC server listening on %s", lis.Addr())
				grpcServer.Serve(lis)
			}()

			// Start the REST proxy for the gRPC server above.
			mux := proxy.NewServeMux()
			for _, r := range proxyRegstr {
				err := r(ctx, mux, cfg.RPCListener.String(), []grpc.DialOption{grpc.WithInsecure()})
				if err != nil {
					return err
				}
			}

			go func() {
				log.Info("REST proxy start listening on %s", cfg.RESTListener.String())
				err := http.ListenAndServe(cfg.RESTListener.String(), mux)
				log.Error("REST proxy failed listening: %s\n", err)
			}()

			return nil
		}()
		close(ce)
		<-stopPoet
		sig.RequestShutdown()
	}()
	return <-ce
}

// loggerInterceptor returns UnaryServerInterceptor handler to log all RPC server incoming requests.
func loggerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		peer, _ := peer.FromContext(ctx)

		if submitReq, ok := req.(*api.SubmitRequest); ok {
			log.Info("%v | %x | %v", info.FullMethod, submitReq.Challenge, peer.Addr.String())
		} else {
			maxDispLen := 50
			reqStr := fmt.Sprintf("%v", req)

			var reqDispStr string
			if len(reqStr) > maxDispLen {
				reqDispStr = reqStr[:maxDispLen] + "..."
			} else {
				reqDispStr = reqStr
			}
			log.Info("%v | %v | %v", info.FullMethod, reqDispStr, peer.Addr.String())
		}

		resp, err := handler(ctx, req)

		if err != nil {
			log.Info("FAILURE %v | %v | %v", info.FullMethod, err, peer.Addr.String())
		}
		return resp, err
	}
}
