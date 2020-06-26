package poet

import (
	"fmt"
	"github.com/spacemeshos/poet/service"
	"net"
	"path/filepath"
	"time"
)

const (
	defaultConfigFilename           = "poet.conf"
	defaultDataDirname              = "data"
	defaultLogDirname               = "logs"
	defaultLogFilename              = "poet.log"
	defaultMaxLogFiles              = 3
	defaultMaxLogFileSize           = 10
	defaultRPCPort                  = 50002
	defaultRESTPort                 = 8080
	defaultN                        = 15
	defaultInitialRoundDuration     = 35 * time.Second
	defaultExecuteEmpty             = true
	defaultMemoryLayers             = 26 // Up to (1 << 26) * 2 - 1 Merkle tree cache nodes (32 bytes each) will be held in-memory
	defaultConnAcksThreshold        = 1
	defaultBroadcastAcksThreshold   = 1
	defaultBroadcastNumRetries      = 100
	defaultBroadcastRetriesInterval = 5 * time.Minute
)

type coreServiceConfig struct {
	N            int  `long:"n" description:"PoET time parameter"`
	MemoryLayers uint `long:"memory" description:"Number of top Merkle tree layers to cache in-memory"`
}

type config struct {
	PoetDir         string `long:"poetdir" description:"The base directory that contains poet's data, logs, configuration file, etc."`
	ConfigFile      string `short:"c" long:"configfile" description:"Path to configuration file"`
	DataDir         string `short:"b" long:"datadir" description:"The directory to store poet's data within"`
	LogDir          string `long:"logdir" description:"Directory to log output."`
	JSONLog         bool   `long:"jsonlog" description:"Whether to log in JSON format"`
	MaxLogFiles     int    `long:"maxlogfiles" description:"Maximum logfiles to keep (0 for no rotation)"`
	MaxLogFileSize  int    `long:"maxlogfilesize" description:"Maximum logfile size in MB"`
	RawRPCListener  string `short:"r" long:"rpclisten" description:"The interface/port/socket to listen for RPC connections"`
	RawRESTListener string `short:"w" long:"restlisten" description:"The interface/port/socket to listen for REST connections"`
	RPCListener     net.Addr
	RESTListener    net.Addr

	CPUProfile string `long:"cpuprofile" description:"Write CPU profile to the specified file"`
	Profile    string `long:"profile" description:"Enable HTTP profiling on given port -- must be between 1024 and 65535"`

	CoreServiceMode bool `long:"core" description:"Enable poet in core service mode"`

	CoreService *coreServiceConfig `group:"Core Service" namespace:"core"`
	Service     *service.Config    `group:"Service"`
}

func PoetDefaultConfig(firstPort int, gatewayPort int, dataDir string) *config {
	poetDir := filepath.Join(dataDir,"poet")
	return &config{
		PoetDir:         poetDir,
		ConfigFile:      filepath.Join(poetDir, defaultConfigFilename),
		DataDir:         filepath.Join(poetDir, defaultDataDirname),
		LogDir:          filepath.Join(poetDir, defaultLogDirname),
		MaxLogFiles:     defaultMaxLogFiles,
		MaxLogFileSize:  defaultMaxLogFileSize,
		RawRESTListener: fmt.Sprintf("localhost:%d", firstPort),
		RawRPCListener:  fmt.Sprintf("localhost:%d", firstPort+1),
		Service: &service.Config{
			N:                        defaultN,
			MemoryLayers:             defaultMemoryLayers,
			InitialRoundDuration:     defaultInitialRoundDuration,
			ExecuteEmpty:             defaultExecuteEmpty,
			ConnAcksThreshold:        defaultConnAcksThreshold,
			BroadcastAcksThreshold:   defaultBroadcastAcksThreshold,
			BroadcastNumRetries:      defaultBroadcastNumRetries,
			BroadcastRetriesInterval: defaultBroadcastRetriesInterval,
			GatewayAddresses: 		  []string{fmt.Sprintf("localhost:%d",gatewayPort)},
		},
		CoreService: &coreServiceConfig{
			N:            defaultN,
			MemoryLayers: defaultMemoryLayers,
		},
	}
}
