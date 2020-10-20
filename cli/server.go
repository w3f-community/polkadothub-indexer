package cli

import (
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/server"
	"github.com/figment-networks/polkadothub-indexer/usecase"
)

func startServer(cfg *config.Config) error {
	client, err := initClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	db, err := initPostgres(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	httpHandlers := usecase.NewHttpHandlers(cfg, client, db.GetAccounts(), db.GetBlocks(), db.GetDatabase(), db.GetEvents(), db.GetReports(), db.GetSyncables(), db.GetValidators())

	a := server.New(cfg, httpHandlers)
	if err := a.Start(cfg.ListenAddr()); err != nil {
		return err
	}
	return nil
}
