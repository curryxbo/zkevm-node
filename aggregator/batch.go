package aggregator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/0xPolygonHermez/zkevm-node/state"
)

func (a *Aggregator) tryGenerateBatchProof(ctx context.Context, prover proverInterface) (bool, error) {
	proverId := prover.ID()
	proverName := prover.Name()

	log.Debugf("tryGenerateBatchProof started, prover: %s, proverId: %s", proverName, proverId)

	handleError := func(batchNumber uint64, logError error) (bool, error) {
		log.Errorf(logError.Error())

		err := a.State.DeleteBatchProofs(a.ctx, batchNumber, batchNumber, nil)
		if err != nil {
			log.Errorf("failed to delete batch %d proof in progress, error: %v", batchNumber, err)
		}

		log.Debugf("tryGenerateBatchProof ended with errors, prover: %s, proverId: %s", proverName, proverId)

		return false, logError
	}

	batchToProve, proof, err := a.getAndLockBatchToProve(ctx, prover)
	if errors.Is(err, state.ErrNotFound) {
		log.Debugf("no batches to prove, prover: %s, proverId: %s", proverName, proverId)
		return false, nil
	}
	if err != nil {
		return false, err
	}

	log.Infof("found batch %d pending to generate proof", batchToProve.BatchNumber)

	inputProver, err := a.buildBatchProverInputs(ctx, batchToProve)
	if err != nil {
		err := fmt.Errorf("failed to build batch %d proof prover request, error: %v", batchToProve.BatchNumber, err)
		return handleError(batchToProve.BatchNumber, err)
	}

	b, err := json.Marshal(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to serialize batch %d proof request, error: %v", batchToProve.BatchNumber, err)
		return handleError(batchToProve.BatchNumber, err)
	}

	proof.InputProver = string(b)

	log.Infof("sending batch %d proof request, prover: %s, proverId: %s", batchToProve.BatchNumber, proverName, proverId)

	proofId, err := prover.BatchProof(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to generate batch %d proof %s, error: %v", batchToProve.BatchNumber, err)
		return handleError(batchToProve.BatchNumber, err)
	}

	proof.Id = proofId

	log.Infof("generating batch %d proof, proofId: %s", batchToProve.BatchNumber, *proof.Id)

	proofData, err := prover.WaitRecursiveProof(ctx, *proof.Id)
	if err != nil {
		err = fmt.Errorf("failed to get batch %d proof from prover, proverId: %s, error: %v", batchToProve.BatchNumber, *proof.Id, err)
		return handleError(batchToProve.BatchNumber, err)
	}

	log.Infof("generated batch %d proof, proofId: %s", batchToProve.BatchNumber, *proof.Id)

	proof.Data = proofData
	proof.GeneratingSince = nil

	err = a.State.UpdateBatchProof(a.ctx, proof, nil)
	if err != nil {
		err = fmt.Errorf("failed to update batch %d proof, error: %v", batchToProve.BatchNumber, err)
		return handleError(batchToProve.BatchNumber, err)
	}

	log.Debugf("tryGenerateBatchProof ended, prover: %s, proverId: %s", proverName, proverId)

	return true, nil
}

func (a *Aggregator) getAndLockBatchToProve(ctx context.Context, prover proverInterface) (*state.Batch, *state.BatchProof, error) {
	proverID := prover.ID()
	proverName := prover.Name()

	a.StateDBMutex.Lock()
	defer a.StateDBMutex.Unlock()

	lastVerifiedBatch, err := a.State.GetLastVerifiedBatch(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	// Calculate max L1 block number for getting next virtual batch to prove
	maxL1BlockNumber, err := a.getMaxL1BlockNumber(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get max L1 block number, error: %v", err)
	}

	log.Debugf("max L1 block number for getting next virtual batch to prove: %d", maxL1BlockNumber)

	// Get virtual batch pending to generate proof
	batchToProve, err := a.State.GetVirtualBatchToProve(ctx, lastVerifiedBatch.BatchNumber, maxL1BlockNumber, nil)
	if err != nil {
		return nil, nil, err
	}

	// pass pol collateral as zero here, bcs in smart contract fee for aggregator is not defined yet
	isProfitable, err := a.ProfitabilityChecker.IsProfitable(ctx, big.NewInt(0))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check aggregator profitability, error: %v", err)
	}

	if !isProfitable {
		log.Infof("batch %d is not profitable, pol collateral %d", batchToProve.BatchNumber, big.NewInt(0))
		//TODO: fix this as err will be nil here
		return nil, nil, err
	}

	now := time.Now().Round(time.Microsecond)
	proof := &state.BatchProof{
		BatchNumber:      batchToProve.BatchNumber,
		BatchNumberFinal: batchToProve.BatchNumber,
		Proof: state.Proof{
			Prover:          &proverName,
			ProverID:        &proverID,
			GeneratingSince: &now,
		},
	}

	// Avoid other prover to process the same batch
	err = a.State.AddBatchProof(ctx, proof, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add batch %d proof to state db, err: %v", batchToProve.BatchNumber, err)
	}

	return batchToProve, proof, nil
}

func (a *Aggregator) tryAggregateBatchProofs(ctx context.Context, prover proverInterface) (bool, error) {
	proverName := prover.Name()
	proverId := prover.ID()

	log.Debug("tryAggregateProofs started, prover: %s, proverId: %s", proverName, proverId)

	handleError := func(proof1 *state.BatchProof, proof2 *state.BatchProof, logError error) (bool, error) {
		log.Errorf(logError.Error())

		err := a.unlockBatchProofsToAggregate(a.ctx, proof1, proof2)
		if err != nil {
			log.Errorf("failed to release aggregated batch proofs %s, err: %v", getAggrBatchTag(proof1, proof2), err)
		}

		log.Debugf("tryAggregateBatchProofs ended with errors, prover: %s, proverId: %s", proverName, proverId)

		return false, logError
	}

	proof1, proof2, err := a.getAndLockBatchProofsToAggregate(ctx, prover)
	if errors.Is(err, state.ErrNotFound) {
		log.Debug("no batch proofs to aggregate")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	aggregationTag := getAggrBatchTag(proof1, proof2)

	log.Infof("aggregating batch proofs %s", aggregationTag)

	inputProver := map[string]interface{}{
		"recursive_proof_1": proof1.Proof,
		"recursive_proof_2": proof2.Proof,
	}
	b, err := json.Marshal(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to serialize aggregate batch prover request, error: %v", err)
		return handleError(proof1, proof2, err)
	}

	proof := &state.BatchProof{
		BatchNumber:      proof1.BatchNumber,
		BatchNumberFinal: proof2.BatchNumberFinal,
		Proof: state.Proof{
			Prover:      &proverName,
			ProverID:    &proverId,
			InputProver: string(b),
		},
	}

	log.Infof("sending aggregated batch %d proof request, prover: %s, proverId: %s", aggregationTag, *proof.Id, proverName, proverId)

	proofId, err := prover.AggregatedBatchProof(proof1.Data, proof2.Data)
	if err != nil {
		err = fmt.Errorf("failed to generate aggregated batch proof %s, error: %v", aggregationTag, err)
		return handleError(proof1, proof2, err)
	}

	proof.Id = proofId

	log.Infof("generating aggregated batch %d proof, proofId: %s, prover: %s, proverId: %s", aggregationTag, *proof.Id, proverName, proverId)

	proofData, err := prover.WaitRecursiveProof(ctx, *proof.Id)
	if err != nil {
		err = fmt.Errorf("failed to get aggregated batch proof from prover, proverId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	log.Info("generated aggregated batch %d proof, proofId: %s", aggregationTag, *proof.Id)

	proof.Data = proofData

	// update the state by removing the 2 aggregated proofs and storing the newly generated recursive proof
	dbTx, err := a.State.BeginStateTransaction(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction to update batch proof aggregation, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	err = a.State.DeleteBatchProofs(ctx, proof1.BatchNumber, proof2.BatchNumberFinal, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback batch proof aggregation, proofId: %s, error: %v", *proof.Id, err)
			return handleError(proof1, proof2, err)
		}
		err = fmt.Errorf("failed to delete previously aggregated batch proofs, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	err = a.State.AddBatchProof(ctx, proof, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback batch proof aggregation, proofId: %s, error: %v", *proof.Id, err)
			return handleError(proof1, proof2, err)
		}
		err = fmt.Errorf("failed to add new batch recursive proof in state db, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		err = fmt.Errorf("failed to commit transaction to update batch proof aggregation, proofId: %s, error: %v", *proof.Id, err)
		return handleError(proof1, proof2, err)
	}

	return true, nil
}

func (a *Aggregator) getAndLockBatchProofsToAggregate(ctx context.Context, prover proverInterface) (*state.BatchProof, *state.BatchProof, error) {
	a.StateDBMutex.Lock()
	defer a.StateDBMutex.Unlock()

	proof1, proof2, err := a.State.GetBatchProofsToAggregate(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	aggregationTag := getAggrBatchTag(proof1, proof2)

	// Set proofs in generating state in a single transaction
	dbTx, err := a.State.BeginStateTransaction(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction to update batch proof aggregation %s, error: %v", aggregationTag, err)
	}

	now := time.Now().Round(time.Microsecond)
	proof1.GeneratingSince = &now
	err = a.State.UpdateBatchProof(ctx, proof1, dbTx)
	if err == nil {
		proof2.GeneratingSince = &now
		err = a.State.UpdateBatchProof(ctx, proof2, dbTx)
	}

	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed to rollback batch proof aggregation %s, error: %v", aggregationTag, err)
		}
		return nil, nil, fmt.Errorf("failed to update batch proof aggregation %s in state db, error: %v", aggregationTag, err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction to update batch proof aggregation %s, error: %v", aggregationTag, err)
	}

	return proof1, proof2, nil
}

func (a *Aggregator) unlockBatchProofsToAggregate(ctx context.Context, proof1 *state.BatchProof, proof2 *state.BatchProof) error {
	aggregationTag := getAggrBatchTag(proof1, proof2)

	// Release proofs from generating state in a single transaction
	dbTx, err := a.State.BeginStateTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction to release batch proof aggregation %s, error: %v", aggregationTag, err)
	}

	proof1.GeneratingSince = nil
	err = a.State.UpdateBatchProof(ctx, proof1, dbTx)
	if err == nil {
		proof2.GeneratingSince = nil
		err = a.State.UpdateBatchProof(ctx, proof2, dbTx)
	}

	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			return fmt.Errorf("failed to rollback release batch proof aggregation %s, error: %v", aggregationTag, err)
		}
		return fmt.Errorf("failed to release batch proof aggregation %s in state db, error: %v", aggregationTag, err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction to release batch proof aggregation %s, error: %v", aggregationTag, err)
	}

	return nil
}

func getAggrBatchTag(proof1 *state.BatchProof, proof2 *state.BatchProof) string {
	return fmt.Sprintf("[%d-%d, %d-%d]", proof1.BatchNumber, proof1.BatchNumberFinal, proof2.BatchNumber, proof2.BatchNumberFinal)
}
