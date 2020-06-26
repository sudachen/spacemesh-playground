package testnet

import (
	"github.com/spacemeshos/go-spacemesh/activation"
	"github.com/spacemeshos/go-spacemesh/config"
)

func TestNetDefaultConfig(firstRestPort, firstRpcPort int, dataDir string) *config.Config {
	cfg := config.DefaultConfig()

	cfg.POST = activation.DefaultConfig()
	cfg.POST.Difficulty = 5
	cfg.POST.NumProvenLabels = 10
	cfg.POST.SpacePerUnit = 1 << 10 // 1KB.
	cfg.POST.NumFiles = 1

	cfg.HARE.N = 5
	cfg.HARE.F = 2
	cfg.HARE.RoundDuration = 3
	cfg.HARE.WakeupDelta = 5
	cfg.HARE.ExpectedLeaders = 5
	cfg.HARE.SuperHare = true
	cfg.LayerAvgSize = 5
	cfg.LayersPerEpoch = 3
	cfg.Hdist = 5

	cfg.LayerDurationSec = 20
	cfg.HareEligibility.ConfidenceParam = 4
	cfg.HareEligibility.EpochOffset = 0
	cfg.StartMining = true
	cfg.SyncRequestTimeout = 2000
	cfg.SyncInterval = 2
	cfg.SyncValidationDelta = 5

	cfg.API.StartJSONServer = true
	cfg.API.JSONServerPort = firstRestPort
	cfg.API.StartGrpcServer = true
	cfg.API.GrpcServerPort = firstRpcPort

	return &cfg
}

