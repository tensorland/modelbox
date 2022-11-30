package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/tensorland/modelbox/server/storage"
	"go.uber.org/zap"
)

type ActionScheduler struct {
	storageIf storage.MetadataStorage
	syncIntv  time.Duration
	stopCh    chan struct{}
	logger    *zap.Logger
}

func NewActionScheduler(storageIf storage.MetadataStorage, intv time.Duration, logger *zap.Logger) *ActionScheduler {
	return &ActionScheduler{
		storageIf: storageIf,
		syncIntv:  intv,
		logger:    logger,
	}
}

func (a *ActionScheduler) Start() error {
	go a.heartBeat()
	return nil
}

func (a *ActionScheduler) Stop() error {
	close(a.stopCh)
	return nil
}

func (a *ActionScheduler) heartBeat() {
	next := time.After(a.syncIntv)
	for {
		select {
		case <-a.stopCh:
			a.logger.Sugar().Infof("stopping action scheduler")
			return
		case <-next:
			a.runScheduler()
			next = time.After(a.syncIntv)
		}
	}
}

func (a *ActionScheduler) runScheduler() error {
	evals, err := a.storageIf.GetActionEvals(context.Background())
	if err != nil {
		return fmt.Errorf("unable to get action evals: %v", err)
	}
	for _, eval := range evals {
		// Handle create evals
		switch eval.ParentType {
		case storage.EvalParentAction:
			return a.handleActionEvals(eval)
		default:
			return fmt.Errorf("unknown eval parent: %v", eval.ParentType)
		}
	}
	return nil
}

func (a *ActionScheduler) handleActionEvals(eval *storage.ActionEval) error {
	switch eval.Type {
	case storage.EvalTypeActionCreated:
		return a.handleCreateActionEval(eval)
	default:
		return fmt.Errorf("unknown eval type: %v", eval.Type)
	}
}

func (a *ActionScheduler) UpdateInstanceStatus(ctx context.Context, update *storage.ActionInstanceUpdate) error {
	ai, err := a.storageIf.GetActionInstance(ctx, update.ActionInstanceId)
	if err != nil {
		return err
	}

	hasUpdated, eval := ai.Update(update)
	if !hasUpdated {
		return nil
	}
	return a.storageIf.UpdateActionInstance(ctx, ai, eval)
}

func (a *ActionScheduler) handleCreateActionEval(eval *storage.ActionEval) error {
	actionState, err := a.storageIf.GetAction(context.Background(), eval.ParentId)
	if err != nil {
		return fmt.Errorf("unable to get action with id: %v", err)
	}
	if len(actionState.Instances) > 0 {
		return nil
	}
	actionInstance := storage.NewActionInstance(eval.ParentId, 0)
	if err := a.storageIf.CreateActionInstance(context.Background(), actionInstance, eval); err != nil {
		return fmt.Errorf("unable to create action instance: %v", err)
	}
	return nil
}
