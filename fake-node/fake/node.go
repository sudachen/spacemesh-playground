package fake

import (
	"github.com/spacemeshos/amcl"
	"github.com/spacemeshos/amcl/BLS381"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/hare"
	harecfg "github.com/spacemeshos/go-spacemesh/hare/config"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/p2p/service"
	"github.com/spacemeshos/go-spacemesh/signing"
	"github.com/sudachen/spacemesh-playground/errstr"
	"os"
	"path/filepath"
	"time"
)

var genesisTime = time.Now()
const LayerDuration = 10*time.Second
const BlockCacheSize = 20
const LayerSize = 10
const LayersPerEpoch = 10
const Hdist = 5
const AtxsPerBlock = 10
const GenesisActiveSet = 5
const SyncRequestTimeout = 2000
const SyncInterval = 10
const SyncValidationDelta = 30


var hareConfig = harecfg.Config{10, 5, 2, 10, 5, false, 1000, 5}

func Start(rootDir string, restPort int) error {
	_ = os.RemoveAll(rootDir)
	if err := os.MkdirAll(rootDir, 0777); err != nil {
		return errstr.Wrapf(1, err, "failed to create root app folder: %v", err.Error())
	}

	sim := service.NewSimulator()
	n := make([]*fakeNode,2)
	for i := range n {
		x := &fakeNode{}
		if err := x.init("", sim); err != nil {
			return err
		}
		n[i] = x
	}

	return nil
}

type fakeNode struct {
	closers    []interface{}
	id         types.NodeID
	dir        string
	edSgn      *signing.EdSigner
	log        log.Log
	clock      TickProvider
	vrfSigner  *BLS381.BlsSigner
	hare       *hare.Hare
	swarm      *service.Node
	store      *store
}

type network interface {
	NewNode() *service.Node
}

func (n *fakeNode) init(rootDir string, net network) (err error) {
	n.edSgn = signing.NewEdSigner()
	rng := amcl.NewRAND()
	pub := n.edSgn.PublicKey().Bytes()
	rng.Seed(len(pub), n.edSgn.Sign(pub)) // assuming ed.private is random, the sig can be used as seed
	vrfPriv, vrfPub := BLS381.GenKeyPair(rng)
	n.vrfSigner = BLS381.NewBlsSigner(vrfPriv)
	n.id = types.NodeID{Key: n.edSgn.PublicKey().String(), VRFPublicKey: vrfPub}
	n.log = log.NewDefault(n.id.ShortString())
	//n.clock = timesync.NewClock(timesync.RealClock{}, LayerDuration, genesisTime, n.log)
	n.clock = newManualClock(genesisTime)
	n.dir = filepath.Join(rootDir,n.id.String())
	n.swarm = net.NewNode()
	n.store, err = newStore(filepath.Join(n.dir,"db"), n.log)
	if err != nil { return }
	n.hare = n.store.synchare(hareConfig, n.swarm, n.edSgn, n.id, n.clock)
	return nil
}
