package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/tensorland/modelbox/server/storage"
	"go.uber.org/zap"
)

type ActionScheduler struct {
	storageIf       storage.MetadataStorage
	syncIntv        time.Duration
	agentHbDeadline time.Duration
	stopCh          chan struct{}
	logger          *zap.Logger
}

func NewActionScheduler(storageIf storage.MetadataStorage, intv time.Duration, logger *zap.Logger) *ActionScheduler {
	// TODO Make agent hb deadline configurable
	return &ActionScheduler{
		storageIf:       storageIf,
		syncIntv:        intv,
		agentHbDeadline: 10 * time.Second,
		stopCh:          make(chan struct{}),
		logger:          logger,
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
	a.logger.Sugar().Infof("[scheduler] starting the scheduler in duration: %v", a.syncIntv)
	next := time.After(a.syncIntv)
	for {
		select {
		case <-a.stopCh:
			a.logger.Sugar().Infof("[scheduler] stopping action scheduler")
			return
		case <-next:
			a.runScheduler()
			next = time.After(a.syncIntv)
		}
	}
}

func (a *ActionScheduler) runScheduler() error {
	ctx := context.Background()
	events, err := a.storageIf.GetUnprocessedChangeEvents(ctx)
	if err != nil {
		return err
	}
	for _, event := range events {
		triggers, err := a.storageIf.GetTriggers(ctx, event.ObjectId)
		if err != nil {
			return err
		}
		switch event.EventType {
		case storage.EventTypeExperimentCreated:
			fallthrough
		case storage.EventTypeModelCreated:
			fallthrough
		case storage.EventTypeModelVersionCreated:
			if err = a.evaluateTrigger(ctx, triggers, event); err != nil {
				return a.evaluateTrigger(ctx, triggers, event)
			}

		case storage.EventTypeActionCreated:
			if err := a.handleActionCreatedEvent(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *ActionScheduler) evaluateTrigger(ctx context.Context, triggers []*storage.Trigger, event *storage.ChangeEvent) error {
	for _, trigger := range triggers {
		action := trigger.GetAction(event)
		if err := a.storageIf.CreateAction(ctx, action); err != nil {
			return err
		}
	}
	return nil
}

func (a *ActionScheduler) evictDeadAgents(ctx context.Context) error {
	_, err := a.storageIf.GetDeadAgents(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *ActionScheduler) handleActionCreatedEvent(ctx context.Context, event *storage.ChangeEvent) error {
	actionInstance := storage.NewActionInstance(event.Action.Id, 0)
	if err := a.storageIf.CreateActionInstance(context.Background(), actionInstance, event); err != nil {
		return fmt.Errorf("[scheduler] unable to create action instance: %v", err)
	}
	return nil
}

func (a *ActionScheduler) UpdateInstanceStatus(ctx context.Context, update *storage.ActionInstanceUpdate) (bool, error) {
	ai, err := a.storageIf.GetActionInstance(ctx, update.ActionInstanceId)
	if err != nil {
		return false, err
	}

	hasUpdated := ai.Update(update)
	if !hasUpdated {
		return false, nil
	}
	if err := a.storageIf.UpdateActionInstance(ctx, ai); err != nil {
		return false, err
	}
	return true, nil
}

func (a *ActionScheduler) GetRunnableActions(ctx context.Context, arch, action string) ([]*storage.ActionInstance, error) {
	instances, err := a.storageIf.GetActionInstances(ctx, storage.StatusPending)
	if err != nil {
		return nil, err
	}

	var runnableInstances []*storage.ActionInstance

	for _, instance := range instances {
		action, err := a.storageIf.GetAction(ctx, instance.ActionId)
		if err != nil {
			return nil, err
		}
		if action.Action.Arch == arch {
			runnableInstances = append(runnableInstances, instance)
		}
	}
	return runnableInstances, nil
}
