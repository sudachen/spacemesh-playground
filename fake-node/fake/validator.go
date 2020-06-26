package fake

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/signing"
)

type validator struct {}

func (validator) Validate(minerID signing.PublicKey, nipst *types.NIPST, expectedChallenge types.Hash32) error {
	return nil
}

func (validator) VerifyPost(minerID signing.PublicKey, proof *types.PostProof, space uint64) error {
	return nil
}
