package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

const (
	ValidatorAggCreatorTaskName = "ValidatorAggCreator"
)

var (
	_ pipeline.Task = (*validatorAggCreatorTask)(nil)
)

func NewValidatorAggCreatorTask(validatorAggDb store.ValidatorAgg) *validatorAggCreatorTask {
	return &validatorAggCreatorTask{
		validatorAggDb: validatorAggDb,
	}
}

type validatorAggCreatorTask struct {
	validatorAggDb store.ValidatorAgg
}

func (t *validatorAggCreatorTask) GetName() string {
	return ValidatorAggCreatorTaskName
}

func (t *validatorAggCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())

	payload := p.(*payload)

	parsedValidators := payload.ParsedValidators

	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageAggregator, t.GetName(), payload.CurrentHeight))

	var newValidatorAggs []model.ValidatorAgg
	var updatedValidatorAggs []model.ValidatorAgg
	for stashAccount, validatorData := range parsedValidators {
		existing, err := t.validatorAggDb.FindAggByStashAccount(stashAccount)
		if err != nil {
			if err != store.ErrNotFound {
				return err
			}
			// Create new

			validator := model.ValidatorAgg{
				Aggregate: &model.Aggregate{
					StartedAtHeight: payload.Syncable.Height,
					StartedAt:       payload.Syncable.Time,
					RecentAtHeight:  payload.Syncable.Height,
					RecentAt:        payload.Syncable.Time,
				},

				StashAccount:            stashAccount,
				DisplayName:             validatorData.DisplayName,
				RecentAsValidatorHeight: payload.Syncable.Height,
			}

			if payload.Syncable.LastInSession {
				if validatorData.Performance.GetOnline() {
					validator.AccumulatedUptime = 1
				} else {
					validator.AccumulatedUptime = 0
				}
				validator.AccumulatedUptimeCount = 1
			}

			newValidatorAggs = append(newValidatorAggs, validator)
			continue
		}
		// Update
		validator := &model.ValidatorAgg{
			Aggregate: &model.Aggregate{
				RecentAtHeight: payload.Syncable.Height,
				RecentAt:       payload.Syncable.Time,
			},

			RecentAsValidatorHeight: payload.Syncable.Height,
			AccumulatedUptime:       existing.AccumulatedUptime,
			AccumulatedUptimeCount:  existing.AccumulatedUptimeCount,
			DisplayName:             validatorData.DisplayName,
		}

		if payload.Syncable.LastInSession {
			if validatorData.Performance.GetOnline() {
				validator.AccumulatedUptime++
			}
			validator.AccumulatedUptimeCount++
		}

		existing.Update(validator)

		updatedValidatorAggs = append(updatedValidatorAggs, *existing)

	}
	payload.NewValidatorAggregates = newValidatorAggs
	payload.UpdatedValidatorAggregates = updatedValidatorAggs

	return nil
}
