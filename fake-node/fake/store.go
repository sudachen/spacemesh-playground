package fake

import (
	"github.com/spacemeshos/amcl/BLS381"
	"github.com/spacemeshos/go-spacemesh/activation"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/database"
	"github.com/spacemeshos/go-spacemesh/hare"
	harecfg "github.com/spacemeshos/go-spacemesh/hare/config"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/mesh"
	"github.com/spacemeshos/go-spacemesh/miner"
	"github.com/spacemeshos/go-spacemesh/oracle"
	"github.com/spacemeshos/go-spacemesh/p2p/service"
	"github.com/spacemeshos/go-spacemesh/pendingtxs"
	"github.com/spacemeshos/go-spacemesh/signing"
	"github.com/spacemeshos/go-spacemesh/state"
	"github.com/spacemeshos/go-spacemesh/sync"
	"github.com/spacemeshos/go-spacemesh/tortoise"
	"github.com/sudachen/spacemesh-playground/errstr"
	"os"
	"path/filepath"
	"time"
)

type store struct {
	dbDir 		   string
	db             *database.LDBDatabase
	txPool         *miner.TxMempool
	atxPool		   *miner.AtxMemPool
	mdb		       *mesh.DB
	msh            *mesh.Mesh
	idStore        *activation.IdentityStore
	poetDb         *activation.PoetDb
	atxdb          *activation.DB
	processor      *state.TransactionProcessor
	lg             log.Log
	closers        []interface{}
}

func newStore(dbDir string, lg log.Log) (n *store, err error) {
	n = &store{dbDir: dbDir, lg: lg}
	err = n.init()
	return
}

func (n *store) init() (err error) {
	if err = os.MkdirAll(n.dbDir, 0777); err != nil {
		return errstr.Wrapf(1, err, "failed to create folder '%v': %v", n.dbDir, err.Error())
	}

	n.db, err = database.NewLDBDatabase(filepath.Join(n.dbDir, "state"), 0, 0, n.lg)
	if err != nil {
		return errstr.Wrapf(1, err, "failed to create database (state): %v", err.Error())
	}
	atxdbstore, err := database.NewLDBDatabase(filepath.Join(n.dbDir, "atx"), 0, 0, n.lg)
	if err != nil {
		return errstr.Wrapf(1, err, "failed to create database (atx): %v", err.Error())
	}
	iddbstore, err := database.NewLDBDatabase(filepath.Join(n.dbDir, "ids"), 0, 0, n.lg)
	if err != nil {
		return errstr.Wrapf(1, err, "failed to create database (ids): %v", err.Error())
	}
	appliedTxs, err := database.NewLDBDatabase(filepath.Join(n.dbDir, "appliedTxs"), 0, 0, n.lg)
	if err != nil {
		return errstr.Wrapf(1, err, "failed to create database (appliedTxs): %v", err.Error())
	}

	n.mdb, err = mesh.NewPersistentMeshDB(filepath.Join(n.dbDir, "mesh"), BlockCacheSize, n.lg)
	if err != nil {
		return errstr.Wrapf(1, err, "failed to create database (mesh): %v", err.Error())
	}

	n.idStore = activation.NewIdentityStore(iddbstore)
	n.atxdb = activation.NewDB(atxdbstore, n.idStore, n.mdb, LayersPerEpoch, validator{}, n.lg)
	trtl := tortoise.NewTortoise(LayerSize, n.mdb, Hdist, n.lg)

	poetDbStore, err := database.NewLDBDatabase(filepath.Join(n.dbDir, "poet"), 0, 0, n.lg)
	if err != nil {
		return err
	}
	n.closers = append(n.closers, poetDbStore)
	n.poetDb = activation.NewPoetDb(poetDbStore, n.lg)

	n.txPool = miner.NewTxMemPool()
	meshAndPoolProjector := pendingtxs.NewMeshAndPoolProjector(n.mdb, n.txPool)
	n.processor = state.NewTransactionProcessor(n.db, appliedTxs, meshAndPoolProjector, n.lg)
	n.atxPool = miner.NewAtxMemPool()
	n.msh = mesh.NewMesh(n.mdb, n.atxdb, mesh.DefaultMeshConfig(), trtl, n.txPool, n.atxPool, n.processor, n.lg)

	return n.genesis()
}

func (n *store) genesis() (err error) {
	return
}

func (n *store)	validationFunc(ids []types.BlockID) bool {
	for _, b := range ids {
		res, err := n.mdb.GetBlock(b)
		if err != nil {
			n.lg.With().Error("output set block not in database", log.BlockID(b.String()), log.Err(err))
			return false
		}
		if res == nil {
			n.lg.With().Error("output set block not in database (BUG BUG BUG - GetBlock return err nil and res nil)", log.BlockID(b.String()))
			return false
		}
	}
	return true
}

func (n *store) synchare(cfg harecfg.Config, swarm *service.Node, edSgn *signing.EdSigner, id types.NodeID, clock TickProvider) *hare.Hare {
	syncConf := sync.Configuration{Concurrency: 4,
		LayerSize:       int(LayerSize),
		LayersPerEpoch:  LayersPerEpoch,
		RequestTimeout:  time.Duration(SyncRequestTimeout) * time.Millisecond,
		SyncInterval:    time.Duration(SyncInterval) * time.Second,
		ValidationDelta: time.Duration(SyncValidationDelta) * time.Second,
		Hdist:           Hdist,
		AtxsLimit:       AtxsPerBlock}
	beaconProvider := &oracle.EpochBeaconProvider{}
	validator := oracle.NewBlockEligibilityValidator(LayerSize, uint32(GenesisActiveSet), LayersPerEpoch, n.atxdb, beaconProvider, BLS381.Verify2, n.lg)
	syncer := sync.NewSync(swarm, n.msh, n.txPool, n.atxPool, validator, n.poetDb, syncConf, clock, n.lg)
	return hare.New(hareConfig, swarm, edSgn, id, n.validationFunc,
		syncer.IsSynced, n.msh, orca{},
		uint16(LayersPerEpoch), n.idStore, orca{},
		clock.Subscribe(), n.lg)
}
