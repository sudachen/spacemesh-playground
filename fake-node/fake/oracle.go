package fake

import (
	"encoding/binary"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"hash/fnv"
)

type orca struct {}

func (orca) IsIdentityActiveOnConsensusView(string, types.LayerID) (bool, error) {
	return true, nil
}

func (orca) Register(isHonest bool, pubkey string) {}

func (orca) Eligible(layer types.LayerID, round int32, committeeSize int, id types.NodeID, sig []byte) (bool, error) {
	return true, nil
}

func (orca) Proof(layer types.LayerID, round int32) ([]byte, error) {
	kInBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(kInBytes, uint32(round))
	hash := fnv.New32()
	_, err := hash.Write(kInBytes)
	if err != nil {
		log.Error("Error writing hash err: %v", err)
	}
	hashBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(hashBytes, uint32(hash.Sum32()))
	return hashBytes, nil
}
