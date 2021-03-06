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
	ErrRunningSequentialReindex = errors.New("indexing skipped because sequential reindex hasn't finished yet")
)

type startUseCase struct {
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

func NewStartUseCase(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	rewardDb store.Rewards, syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) *startUseCase {
	return &startUseCase{
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

func (uc *startUseCase) Execute(ctx context.Context, batchSize int64) error {
	if err := uc.canExecute(); err != nil {
		return err
	}

	indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.client, uc.accountDb, uc.blockDb, uc.databaseDb, uc.eventDb, uc.reportDb, uc.rewardDb, uc.syncableDb, uc.systemEventDb, uc.transactionDb, uc.validatorDb)
	if err != nil {
		return err
	}

	return indexingPipeline.Start(ctx, indexer.IndexConfig{
		BatchSize: batchSize,
	})
}

// canExecute checks if sequential reindex is already running
// if is it running we skip indexing
func (uc *startUseCase) canExecute() error {
	if _, err := uc.reportDb.FindNotCompletedByKind(model.ReportKindSequentialReindex); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	return ErrRunningSequentialReindex
}
