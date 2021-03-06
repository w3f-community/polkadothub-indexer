package indexing

import (
	"context"
	"errors"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
)

var (
	ErrBackfillRunning = errors.New("backfill already running (use force flag to override it)")
)

type backfillUseCase struct {
	cfg    *config.Config
	client *client.Client

	accountDb     store.Accounts
	blockDb       store.Blocks
	databaseDb    store.Database
	eventDb       store.Events
	reportDb      store.Reports
	rewardDb      store.Rewards
	syncableDb    store.Syncables
	systemEventDb store.SystemEvents
	transactionDb store.Transactions
	validatorDb   store.Validators
}

func NewBackfillUseCase(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events,
	reportDb store.Reports, rewardDb store.Rewards, syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) *backfillUseCase {
	return &backfillUseCase{
		cfg:    cfg,
		client: cli,

		accountDb:     accountDb,
		blockDb:       blockDb,
		databaseDb:    databaseDb,
		eventDb:       eventDb,
		reportDb:      reportDb,
		rewardDb:      rewardDb,
		syncableDb:    syncableDb,
		systemEventDb: systemEventDb,
		transactionDb: transactionDb,
		validatorDb:   validatorDb,
	}
}

type BackfillUseCaseConfig struct {
	Parallel  bool
	Force     bool
	TargetIds []int64
}

func (uc *backfillUseCase) Execute(ctx context.Context, useCaseConfig BackfillUseCaseConfig) error {
	if err := uc.canExecute(); err != nil {
		return err
	}

	indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.client, uc.accountDb, uc.blockDb, uc.databaseDb, uc.eventDb, uc.reportDb, uc.rewardDb, uc.syncableDb, uc.systemEventDb, uc.transactionDb, uc.validatorDb)
	if err != nil {
		return err
	}

	return indexingPipeline.Backfill(ctx, indexer.BackfillConfig{
		Parallel:  useCaseConfig.Parallel,
		Force:     useCaseConfig.Force,
		TargetIds: useCaseConfig.TargetIds,
	})
}

// canExecute checks if reindex is already running
// if is it running we skip indexing
func (uc *backfillUseCase) canExecute() error {
	if _, err := uc.reportDb.FindNotCompletedByKind(model.ReportKindSequentialReindex, model.ReportKindParallelReindex); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	return ErrBackfillRunning
}
