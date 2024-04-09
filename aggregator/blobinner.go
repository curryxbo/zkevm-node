package aggregator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/0xPolygonHermez/zkevm-node/state"
)

func (a *Aggregator) tryGenerateBlobInnerProof(ctx context.Context, prover proverInterface) (bool, error) {
	proverId := prover.ID()
	proverName := prover.Name()

	log.Debug("tryGenerateBlobInnerProof started, prover: %s, proverId: %s", proverName, proverId)

	handleError := func(blobInnerNum uint64, logError error) (bool, error) {
		log.Errorf(logError.Error())

		err := a.State.DeleteBlobInnerProof(a.ctx, blobInnerNum, nil)
		if err != nil {
			log.Errorf("failed to delete blobInner %d proof in progress, error: %v", blobInnerNum, err)
		}

		log.Debugf("tryGenerateBlobInnerProof ended with errors, prover: %s, proverId: %s", proverName, proverId)

		return false, logError
	}

	blobInnerToProve, proof, err := a.getAndLockBlobInnerToProve(ctx, prover)
	if errors.Is(err, state.ErrNotFound) {
		log.Debugf("no blobInner to prove, prover: %s, proverId: %s", proverName, proverId)
		return false, nil
	}
	if err != nil {
		return false, err
	}

	log.Infof("found blobInner %d pending to generate proof", blobInnerToProve)

	inputProver, err := a.buildBlobInnerProverInputs(ctx, blobInnerToProve)
	if err != nil {
		err := fmt.Errorf("failed to build blobInner %d proof prover request, error: %v", blobInnerToProve.BlobInnerNum, err)
		return handleError(blobInnerToProve.BlobInnerNum, err)
	}

	b, err := json.Marshal(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to serialize blobInner %d proof request, error: %v", blobInnerToProve.BlobInnerNum, err)
		return handleError(blobInnerToProve.BlobInnerNum, err)
	}

	proof.InputProver = string(b)

	log.Infof("sending batch %d proof request, prover: %s, proverId: %s", blobInnerToProve.BlobInnerNum, proverName, proverId)

	proofId, err := prover.BlobInnerProof(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to get blobInner %d proof from prover, proverId: %s, error: %v", blobInnerToProve.BlobInnerNum, *proof.Id, err)
		return handleError(blobInnerToProve.BlobInnerNum, err)
	}

	proof.Id = proofId

	log.Infof("generating blobInner %d proof, proofId: %s", blobInnerToProve.BlobInnerNum, *proof.Id)

	proofData, err := prover.WaitRecursiveProof(ctx, *proof.Id)
	if err != nil {
		err = fmt.Errorf("failed to get blobInner proof id %s from prover, error: %v", proof.Id, err)
		return handleError(blobInnerToProve.BlobInnerNum, err)
	}

	log.Info("generated blobInner %d proof, Id: %s", blobInnerToProve.BlobInnerNum, *proof.Id)

	proof.Data = proofData
	proof.GeneratingSince = nil

	err = a.State.UpdateBlobInnerProof(a.ctx, proof, nil)
	if err != nil {
		err = fmt.Errorf("failed to update blobInner %d proof, error: %v", blobInnerToProve.BlobInnerNum, err)
		return handleError(blobInnerToProve.BlobInnerNum, err)
	}

	log.Debug("tryGenerateBlobInnerProof ended")

	return true, nil
}

func (a *Aggregator) getAndLockBlobInnerToProve(ctx context.Context, prover proverInterface) (*state.BlobInner, *state.BlobInnerProof, error) {
	proverID := prover.ID()
	proverName := prover.Name()

	a.StateDBMutex.Lock()
	defer a.StateDBMutex.Unlock()

	lastVerifiedBlobInner, err := a.State.GetLastVerifiedBlobInner(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	// Get max L1 block number for getting next blobInner to prove
	maxL1BlockNumber, err := a.getMaxL1BlockNumber(ctx)
	if err != nil {
		log.Errorf("failed to get max L1 block number, error: %v", err)
		return nil, nil, err
	}

	log.Debugf("max L1 block number for getting next blobInner to prove: %d", maxL1BlockNumber)

	// Get blobInner pending to generate proof
	blobInnerToProve, err := a.State.GetBlobInnerToProve(ctx, lastVerifiedBlobInner.BlobInnerNum, maxL1BlockNumber, nil)
	if err != nil {
		return nil, nil, err
	}

	log.Infof("found blobInner %d pending to generate proof", lastVerifiedBlobInner.BlobInnerNum)

	now := time.Now().Round(time.Microsecond)
	proof := &state.BlobInnerProof{
		BlobInnerNumber: blobInnerToProve.BlobInnerNum,
		Proof: state.Proof{
			Prover:          &proverName,
			ProverID:        &proverID,
			GeneratingSince: &now,
		},
	}

	// Avoid other prover to process the same blobInner
	err = a.State.AddBlobInnerProof(ctx, proof, nil)
	if err != nil {
		log.Errorf("failed to add blobInner %d proof, error: %v", proof.BlobInnerNumber, err)
		return nil, nil, err
	}

	return blobInnerToProve, proof, nil
}
