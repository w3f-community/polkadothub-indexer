package validator

import (
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

type getByHeightUseCase struct {
	syncablesDb           store.Syncables
	validatorEraSeqDb     store.ValidatorEraSeq
	validatorSessionSeqDb store.ValidatorSessionSeq
}

func NewGetByHeightUseCase(syncablesDb store.Syncables,
	validatorEraSeqDb store.ValidatorEraSeq, validatorSessionSeqDb store.ValidatorSessionSeq,
) *getByHeightUseCase {
	return &getByHeightUseCase{
		syncablesDb:           syncablesDb,
		validatorEraSeqDb:     validatorEraSeqDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
	}
}

func (uc *getByHeightUseCase) Execute(height *int64) (*SeqListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.syncablesDb.FindMostRecent()
	if err != nil {
		return nil, err
	}
	lastH := mostRecentSynced.Height

	// Show last synced height, if not provided
	if height == nil {
		height = &lastH
	}

	if *height > lastH {
		return nil, errors.New("height is not indexed yet")
	}

	validatorSessionSequences, err := uc.validatorSessionSeqDb.FindByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	validatorEraSequences, err := uc.validatorEraSeqDb.FindByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	return ToSeqListView(validatorSessionSequences, validatorEraSequences), nil
}
