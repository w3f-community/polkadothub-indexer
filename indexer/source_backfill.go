package indexer

import (
	"context"
	"errors"
	"fmt"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
)

var (
	_ pipeline.Source = (*backfillSource)(nil)
)

func NewBackfillSource(cfg *config.Config, syncablesDb store.Syncables, client *client.Client, indexVersion int64, isLastInSession, isLastInEra bool) (*backfillSource, error) {
	src := &backfillSource{
		cfg:         cfg,
		syncablesDb: syncablesDb,
		client:      client,

		currentIndexVersion: indexVersion,
	}

	if err := src.init(isLastInSession, isLastInEra); err != nil {
		return nil, err
	}

	return src, nil
}

type backfillSource struct {
	cfg                 *config.Config
	syncablesDb         store.Syncables
	client              *client.Client
	useWhiteList        bool
	heightsWhitelist    map[int64]int64
	whiteListStages     []pipeline.StageName
	currentIndexVersion int64
	currentHeight       int64
	startHeight         int64
	endHeight           int64
	err                 error
}

func (s *backfillSource) Next(context.Context, pipeline.Payload) bool {
	if s.err == nil && s.currentHeight < s.endHeight {
		s.currentHeight = s.currentHeight + 1
		return true
	}
	return false
}

func (s *backfillSource) isCurrentHeightInWhitelist() bool {
	_, found := s.heightsWhitelist[s.currentHeight]
	return found
}

func (s *backfillSource) UseWhiteList() bool {
	return s.useWhiteList
}

func (s *backfillSource) Current() int64 {
	return s.currentHeight
}

func (s *backfillSource) Err() error {
	return s.err
}

func (s *backfillSource) Skip(stageName pipeline.StageName) bool {
	if s.UseWhiteList() {
		if !s.isCurrentHeightInWhitelist() {
			return !s.isStageInWhiteList(stageName)
		} else {
			return false
		}
	}

	return false
}

func (s *backfillSource) Len() int64 {
	return s.endHeight - s.startHeight + 1
}

func (s *backfillSource) init(isLastInSession, isLastInEra bool) error {
	s.useWhiteList = isLastInSession || isLastInEra
	if s.UseWhiteList() {
		if err := s.setHeightsWhitelist(isLastInSession, isLastInEra); err != nil {
			return err
		}
	}
	if err := s.setStartHeight(); err != nil {
		return err
	}
	if err := s.setEndHeight(); err != nil {
		return err
	}
	return nil
}

func (s *backfillSource) setStartHeight() error {
	syncable, err := s.syncablesDb.FindFirstByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.currentHeight = syncable.Height
	s.startHeight = syncable.Height
	return nil
}

func (s *backfillSource) setEndHeight() error {
	syncable, err := s.syncablesDb.FindMostRecentByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.endHeight = syncable.Height
	return nil
}

func (s *backfillSource) setHeightsWhitelist(isLastInSession, isLastInEra bool) error {
	syncables, err := s.syncablesDb.FindAllByLastInSessionOrEra(s.currentIndexVersion, isLastInSession, isLastInEra)
	if err != nil {
		return err
	}
	if len(syncables) == 0 {
		return errors.New(fmt.Sprintf("no heights for whitelist to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
	}

	s.generateMapForWhiteList(syncables)
	return nil
}

func (s *backfillSource) generateMapForWhiteList(syncables []model.Syncable) {
	heightsWhitelist := make(map[int64]int64)
	for i := 0; i < len(syncables); i++ {
		heightsWhitelist[syncables[i].Height] = syncables[i].Height
	}
	s.heightsWhitelist = heightsWhitelist
}

func (s *backfillSource) isStageInWhiteList(stageName pipeline.StageName) bool {
	if pipeline.StageSyncer == stageName {
		return true
	}

	return false
}
