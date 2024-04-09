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

func (a *Aggregator) tryGenerateBlobOuterProof(ctx context.Context, prover proverInterface) (bool, error) {
	proverId := prover.ID()
	proverName := prover.Name()

	log.Debug("tryGenerateBlobOuterProof started, prover: %s, proverId: %s", proverName, proverId)

	handleError := func(blobOuterNumber uint64, logError error) (bool, error) {
		log.Errorf(logError.Error())

		err := a.State.DeleteBlobOuterProofs(a.ctx, blobOuterNumber, blobOuterNumber, nil)
		if err != nil {
			log.Errorf("failed to delete blobOuter %d proof in progress, error: %v", blobOuterNumber, err)
		}

		log.Debug("tryGenerateBlobOuterProof ended with errors")

		return false, logError
	}

	batchProof, blobInnerProof, blobOuterProof, err := a.getAndLockBlobOuterToProve(ctx, prover)
	if errors.Is(err, state.ErrNotFound) {
		log.Debug("no blobOuter pending to prove")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	log.Infof("found blobOuter %d (batches[%d-%d]) pending to generate proof", blobInnerProof.BlobInnerNumber, batchProof.BatchNumber, batchProof.BatchNumberFinal)

	inputProver := map[string]interface{}{
		"batch_proof":      batchProof.Data,
		"blob_inner_proof": blobInnerProof.Data,
	}
	b, err := json.Marshal(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to serialize blobOuter %d proof request, error: %v", blobOuterProof.BlobOuterNumber, err)
		return handleError(blobOuterProof.BlobOuterNumber, err)
	}

	blobOuterProof.InputProver = string(b)

	log.Infof("sending blobOuter %d proof request, prover: %s, proverId: %s", blobOuterProof.BlobOuterNumber, proverName, proverId)

	proofId, err := prover.BlobOuterProof(batchProof.Data, blobInnerProof.Data)
	if err != nil {
		err = fmt.Errorf("failed to generate blobOuter %d proof, error: %v", blobOuterProof.BlobOuterNumber, err)
		return handleError(blobOuterProof.BlobOuterNumber, err)
	}

	blobOuterProof.Id = proofId

	log.Infof("generating blobOuter %d proof, proofId: %s", blobOuterProof.BlobOuterNumber, blobOuterProof.Id)

	proofData, err := prover.WaitRecursiveProof(ctx, *blobOuterProof.Id)
	if err != nil {
		err = fmt.Errorf("failed to get blobOuter proof from prover, proofId: %s, error: %v", *blobInnerProof.Id, err)
		return handleError(blobOuterProof.BlobOuterNumber, err)
	}

	log.Info("generated blobOuter %d proof, proofId: %s", blobOuterProof.BlobOuterNumber, blobOuterProof.Id)

	blobOuterProof.Data = proofData
	blobOuterProof.GeneratingSince = nil

	err = a.State.UpdateBlobOuterProof(a.ctx, blobOuterProof, nil)
	if err != nil {
		err = fmt.Errorf("failed to update blobOuter %d proof in state db, proofId: %s, error: %v", blobOuterProof.BlobOuterNumber, *blobOuterProof.Id, err)
		return handleError(blobOuterProof.BlobOuterNumber, err)
	}

	log.Debug("tryGenerateBlobOuterProof ended")

	return true, nil
}

func (a *Aggregator) getAndLockBlobOuterToProve(ctx context.Context, prover proverInterface) (*state.BatchProof, *state.BlobInnerProof, *state.BlobOuterProof, error) {
	proverID := prover.ID()
	proverName := prover.Name()

	a.StateDBMutex.Lock()
	defer a.StateDBMutex.Unlock()

	// Get blobOuter pending to generate proof
	batchProof, blobInnerProof, err := a.State.GetBlobOuterToProve(ctx, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get blobOuter to prove from state db, err: %v", err)
	}

	now := time.Now().Round(time.Microsecond)
	blobOuterProof := &state.BlobOuterProof{
		BlobOuterNumber:      blobInnerProof.BlobInnerNumber,
		BlobOuterNumberFinal: blobInnerProof.BlobInnerNumber,
		Proof: state.Proof{
			Prover:          &proverName,
			ProverID:        &proverID,
			GeneratingSince: &now,
		},
	}

	// Avoid other prover to process the same blobInner
	err = a.State.AddBlobOuterProof(ctx, blobOuterProof, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to add blobOuter %d proof to state db, err: %v", blobOuterProof.BlobOuterNumber, err)
	}

	return batchProof, blobInnerProof, blobOuterProof, nil
}

func (a *Aggregator) tryAggregateBlobOuterProofs(ctx context.Context, prover proverInterface) (bool, error) {
	proverId := prover.ID()
	proverName := prover.Name()

	log.Debug("tryGenerateBlobOuterProof started, prover: %s, proverId: %s", proverName, proverId)

	handleError := func(proof1 *state.BlobOuterProof, proof2 *state.BlobOuterProof, logError error) (bool, error) {
		log.Errorf(logError.Error())

		err := a.unlockBlobOuterProofsToAggregate(a.ctx, proof1, proof2)
		if err != nil {
			log.Errorf("failed to release aggregated blobOuter proofs %s, err: %v", getAggrBlobOuterTag(proof1, proof2), err)
		}

		log.Debug("tryAggregateBlobOuterProofs ended with errors")

		return false, logError
	}

	proof1, proof2, err := a.getAndLockBlobOuterProofsToAggregate(ctx, prover)
	if errors.Is(err, state.ErrNotFound) {
		log.Debug("no blobOuter proofs to aggregate")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	var (
		proofId *string
	)

	aggregationTag := getAggrBlobOuterTag(proof1, proof2)

	log.Infof("aggregating blobOuter proofs %s", aggregationTag)

	inputProver := map[string]interface{}{
		"recursive_proof_1": proof1.Proof,
		"recursive_proof_2": proof2.Proof,
	}
	b, err := json.Marshal(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to serialize aggregate blobOuter prover request, error: %v", err)
		return handleError(proof1, proof2, err)
	}

	proof := &state.BlobOuterProof{
		BlobOuterNumber:      proof1.BlobOuterNumber,
		BlobOuterNumberFinal: proof2.BlobOuterNumberFinal,
		Proof: state.Proof{
			Prover:      &proverName,
			ProverID:    &proverId,
			InputProver: string(b),
		},
	}

	log.Infof("sending aggregated blobOuter %d proof request, prover: %s, proverId: %s", aggregationTag, *proof.Id, proverName, proverId)

	proofId, err = prover.AggregatedBlobOuterProof(proof1.Data, proof2.Data)
	if err != nil {
		err = fmt.Errorf("failed to generate aggregated blobOuter proof %s, error: %v", aggregationTag, err)
		return handleError(proof1, proof2, err)
	}

	proof.Id = proofId

	log.Infof("generating aggregated blobOuter %s proof, proofId: %s, prover: %s, proverId: %s", aggregationTag, *proof.Id, proverName, proverId)

	proofData, err := prover.WaitRecursiveProof(ctx, *proof.Id)
	if err != nil {
		err = fmt.Errorf("failed to get aggregated blobOuter proof from prover, proverId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	log.Infof("generated aggregated blobOuter %d proof, proofId: %s", aggregationTag, *proof.Id)

	proof.Data = proofData

	// update the state by removing the 2 blobOuter proofs and storing the newly generated recursive proof
	dbTx, err := a.State.BeginStateTransaction(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction to update blobOuter proof aggregation, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	err = a.State.DeleteBlobOuterProofs(ctx, proof1.BlobOuterNumber, proof2.BlobOuterNumberFinal, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback blobOuter proof aggregation, proofId: %s, error: %v", *proof.Id, err)
			return handleError(proof1, proof2, err)
		}
		err = fmt.Errorf("failed to delete previously aggregated blobOuter proofs, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	now := time.Now().Round(time.Microsecond)
	proof.GeneratingSince = &now // we lock the new blobOuter proof to check if it can send as final
	err = a.State.AddBlobOuterProof(ctx, proof, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback blobOuter proof aggregation, proofId: %s, error: %v", *proof.Id, err)
			return handleError(proof1, proof2, err)
		}
		err = fmt.Errorf("failed to add new blobOuter recursive proof in state db, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		err = fmt.Errorf("failed to commit transaction to update blobOuter proof aggregation, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}


	generateFinal := false
	if a.canVerifyProof() {
		isFinal, err := a.isFinalProof(ctx, proof)
		if err != nil {
			return false, fmt.Errorf("failed to validate if blobOuter proof %s is final, error: %v", aggregationTag, err)
		}

		if isFinal {
			// state is up to date, check if we can send the final proof using the one just crafted
			finalProofBuilt, err := a.tryBuildFinalProof(ctx, prover, proof)
			if err != nil {
				// just log the error and continue to handle the aggregated proof
				log.Errorf("Failed trying to check if recursive proof can be verified: %v", err)
			}

			if !finalProofBuilt {
				// final proof has not been generated, update the recursive proof (unlock it)
				proof.GeneratingSince = nil
				err := a.State.UpdateBlobOuterProof(a.ctx, proof, nil)
				if err != nil {
					return false, fmt.Errorf("failed to update new blobOuter recursive proof in state db, proofId: %s, error: %v", *proof.Id, err)
				}
			}
		}
	}

	if (!generateFinal) {
		proof.GeneratingSince = nil
		err := a.State.UpdateBlobOuterProof(a.ctx, proof, nil)
		if err != nil {
			return false, fmt.Errorf("failed to update new blobOuter recursive proof in state db, proofId: %s, error: %v", *proof.Id, err)
		}
	}

	return true, nil
}

func (a *Aggregator) getAndLockBlobOuterProofsToAggregate(ctx context.Context, prover proverInterface) (*state.BlobOuterProof, *state.BlobOuterProof, error) {
	a.StateDBMutex.Lock()
	defer a.StateDBMutex.Unlock()

	proof1, proof2, err := a.State.GetBlobOuterProofsToAggregate(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	aggregationTag := getAggrBlobOuterTag(proof1, proof2)

	// Set proofs in generating state in a single transaction
	dbTx, err := a.State.BeginStateTransaction(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction to update blobOuter proof aggregation %s, error: %v", aggregationTag, err)
	}

	now := time.Now().Round(time.Microsecond)
	proof1.GeneratingSince = &now
	err = a.State.UpdateBlobOuterProof(ctx, proof1, dbTx)
	if err == nil {
		proof2.GeneratingSince = &now
		err = a.State.UpdateBlobOuterProof(ctx, proof2, dbTx)
	}

	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed to rollback blobOuter proof aggregation %s, error: %v", aggregationTag, err)
		}
		return nil, nil, fmt.Errorf("failed to update blobOuter proof aggregation %s in state db, error: %v", aggregationTag, err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction to update blobOuter proof aggregation %s, error: %v", aggregationTag, err)
	}

	return proof1, proof2, nil
}

func (a *Aggregator) unlockBlobOuterProofsToAggregate(ctx context.Context, proof1 *state.BlobOuterProof, proof2 *state.BlobOuterProof) error {
	aggregationTag := getAggrBlobOuterTag(proof1, proof2)

	dbTx, err := a.State.BeginStateTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction to release blobOuter proof aggregation %s, error: %v", aggregationTag, err)
	}

	proof1.GeneratingSince = nil
	err = a.State.UpdateBlobOuterProof(ctx, proof1, dbTx)
	if err == nil {
		proof2.GeneratingSince = nil
		err = a.State.UpdateBlobOuterProof(ctx, proof2, dbTx)
	}

	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			return fmt.Errorf("failed to rollback release blobOuter proof aggregation %s, error: %v", aggregationTag, err)
		}
		return fmt.Errorf("failed to release blobOuter proof aggregation %s in state db, error: %v", aggregationTag, err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction to release blobOuter proof aggregation %s, error: %v", aggregationTag, err)
	}

	return nil
}

func getAggrBlobOuterTag(proof1 *state.BlobOuterProof, proof2 *state.BlobOuterProof) string {
	return fmt.Sprintf("[%d-%d, %d-%d]", proof1.BlobOuterNumber, proof1.BlobOuterNumberFinal, proof2.BlobOuterNumber, proof2.BlobOuterNumberFinal)
}
