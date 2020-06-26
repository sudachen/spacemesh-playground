package main

import (
	"fmt"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spf13/cobra"
	"github.com/sudachen/spacemesh-playground/fake-node/fake"
	"github.com/sudachen/spacemesh-playground/local-testnet/testnet"
	"os"
	"runtime"
)

const RestPort = 19190
const DataPath = "./fake-node"

func init() {
	log.DebugMode(true)
	log.AppLog = log.NewDefault("")
}

func main() {

	cmd := &cobra.Command{
		Use:           "local-testnet",
		Short:         fmt.Sprintf("Spacemesh Local-TestNet App %v.%v (https://github.com/sudachen/spacemesh-playground/)", MajorVersion, MinorVersion),
		SilenceErrors: true,
		Run: func (cmd *cobra.Command, args []string) {
			log.Debug("test")
			fake.Start(DataPath, RestPort)
		},
	}

	optx := cmd.PersistentFlags().BoolP("trace", "x", false, "backtrace on panic")
	runtime.GOMAXPROCS(runtime.NumCPU())

	defer func() {
		if !*optx {
			if e := recover(); e != nil {
				_,_ = fmt.Fprintln(os.Stderr, errstr.MessageOf(e))
				os.Exit(1)
			}
		}
	}()

	if err := cmd.Execute(); err != nil {
		panic(errstr.Frame(0, err))
	}
}

