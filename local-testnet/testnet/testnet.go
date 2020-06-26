package testnet

import (
	"context"
	"github.com/spacemeshos/amcl/BLS381"
	"github.com/spacemeshos/go-spacemesh/activation"
	"github.com/spacemeshos/go-spacemesh/collector"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/eligibility"
	"github.com/spacemeshos/go-spacemesh/events"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/p2p/service"
	"github.com/sudachen/spacemesh-playground/local-testnet/testnet/poet"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func Start(
	numOfinstances, layerAvgSize int,
	runTillLayer uint32,
	dataDir string,
	firstPoetPort int,
	firstNodeRest, firstNodeRpc int) {

	cfg := TestNetDefaultConfig(firstNodeRest, firstNodeRpc, dataDir)
	cfg.LayerAvgSize = layerAvgSize
	numOfInstances := numOfinstances
	net := service.NewSimulator()
	_ = os.RemoveAll(filepath.Join(dataDir,"node"))
	_ = os.MkdirAll(filepath.Join(dataDir,"node"),0777)
	path := filepath.Join(dataDir,"node","x")
	genesisTime := time.Now().Add(20 * time.Second).Format(time.RFC3339)

	rolacle := eligibility.New()
	rng := BLS381.DefaultSeed()
	gTime, err := time.Parse(time.RFC3339, genesisTime)
	if err != nil {
		log.Error("cannot parse genesis time %v", err)
	}
	pubsubAddr := "tcp://localhost:56565"
	events.InitializeEventPubsub(pubsubAddr)
	clock := NewManualClock(gTime)

	poetCfg := poet.PoetDefaultConfig(firstPoetPort, firstNodeRpc, dataDir)
	poetClient := activation.NewHTTPPoetClient(context.Background(),poetCfg.RawRESTListener)

	apps := make([]*SpacemeshApp, 0, numOfInstances)
	name := 'a'
	for i := 0; i < numOfInstances; i++ {
		dbStorepath := path + string(name)
		smApp, err := InitSingleInstance(*cfg, i, genesisTime, rng, dbStorepath, rolacle, poetClient, clock, net)
		if err != nil {
			log.Error("cannot run multi node %v", err)
			return
		}
		apps = append(apps, smApp)
		name++
	}

	eventDb := collector.NewMemoryCollector()
	collect := collector.NewCollector(eventDb, pubsubAddr)
	for _, a := range apps {
		a.startServices()
	}
	collect.Start(false)
	ActivateGrpcServer(apps[0])
	ActivateJsonServer(apps[0])

	stopPoet:= make(chan struct{})
	defer close(stopPoet)

	if err = poet.StartServer(poetCfg, stopPoet); err != nil {
		log.Error("cannot run poet service %v", err)
		return
	}

	//startInLayer := 5 // delayed pod will start in this layer
	defer GracefulShutdown(apps)

	timeout := time.After(time.Duration(runTillLayer*60) * time.Second)

	// stickyClientsDone := 0
	startLayer := time.Now()
	clock.Tick()
	errors := 0

	sigc := make(chan os.Signal,1)
	signal.Notify(sigc, os.Interrupt)

loop:
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-sigc:
			log.Info("Stop")
			return
		case <-timeout:
			log.Panic("run timed out", err)
			return
		default:
			if errors > 100 {
				log.Panic("too many errors and retries")
				break loop
			}
			layer := clock.GetCurrentLayer()
			if eventDb.GetBlockCreationDone(layer) < numOfInstances {
				log.Info("blocks done in layer %v: %v", layer, eventDb.GetBlockCreationDone(layer))
				time.Sleep(500 * time.Millisecond)
				errors++
				continue
			}
			log.Info("all miners tried to create block in %v", layer)
			if eventDb.GetNumOfCreatedBlocks(layer)*numOfInstances != eventDb.GetReceivedBlocks(layer) {
				log.Info("finished: %v, block received %v layer %v", eventDb.GetNumOfCreatedBlocks(layer), eventDb.GetReceivedBlocks(layer), layer)
				time.Sleep(500 * time.Millisecond)
				errors++
				continue
			}
			log.Info("all miners got blocks for layer: %v created: %v received: %v", layer, eventDb.GetNumOfCreatedBlocks(layer), eventDb.GetReceivedBlocks(layer))
			epoch := layer.GetEpoch(uint16(cfg.LayersPerEpoch))
			if !(eventDb.GetAtxCreationDone(epoch) >= numOfInstances && eventDb.GetAtxCreationDone(epoch)%numOfInstances == 0) {
				log.Info("atx not created %v in epoch %v, created only %v atxs", numOfInstances-eventDb.GetAtxCreationDone(epoch), epoch, eventDb.GetAtxCreationDone(epoch))
				time.Sleep(500 * time.Millisecond)
				errors++
				continue
			}
			log.Info("all miners finished reading %v atxs, layer %v done in %v", eventDb.GetAtxCreationDone(epoch), layer, time.Since(startLayer))
			for _, atxID := range eventDb.GetCreatedAtx(epoch) {
				if _, found := eventDb.Atxs[atxID]; !found {
					log.Info("atx %v not propagated", atxID)
					errors++
					continue
				}
			}
			errors = 0

			startLayer = time.Now()
			clock.Tick()

			if apps[0].mesh.LatestLayer() >= types.LayerID(runTillLayer) {
				break loop
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
	collect.Stop()
}

