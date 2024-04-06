package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"errors"
	"fmt"

	"go.keploy.io/server/v2/pkg/models"
	"go.keploy.io/server/v2/utils"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"go.keploy.io/server/v2/pkg/graph/model"
)

// TestSets is the resolver for the testSets field.
func (r *queryResolver) TestSets(ctx context.Context) ([]string, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(utils.Emoji + "failed to get Resolver")
		return nil, err
	}

	ctx = context.WithoutCancel(ctx)
	ids, err := r.replay.GetAllTestSetIDs(ctx)
	if err != nil {
		utils.LogError(r.logger, err, "failed to get all test set ids")
		return nil, errors.New("failed to get all test sets")
	}
	r.logger.Debug("test set ids", zap.Strings("ids", ids))
	return ids, nil
}

// StartHooks is the resolver for the startHooks field.
func (r *mutationResolver) StartHooks(ctx context.Context) (*model.TestRunInfo, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(utils.Emoji + "failed to get Resolver")
		return nil, err
	}

	ctx = context.WithoutCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	ctx = context.WithValue(ctx, models.ErrGroupKey, g)
	r.hookCtx = ctx

	testRunId, appId, hookCancel, err := r.replay.BootReplay(ctx)
	if err != nil {
		utils.LogError(r.logger, err, "failed to boot replay")
		return nil, errors.New("failed to hook the application")
	}
	r.hookCancel = hookCancel
	r.logger.Debug("test run info", zap.String("testRunId", testRunId), zap.Int("appId", int(appId)))
	return &model.TestRunInfo{
		TestRunID: testRunId,
		AppID:     int(appId),
	}, nil
}

// RunTestSet is the resolver for the runTestSet field.
func (r *mutationResolver) RunTestSet(ctx context.Context, testSetID string, testRunID string, appID int) (bool, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(utils.Emoji + "failed to get Resolver")
		return false, err
	}
	r.logger.Debug("running test set", zap.String("testSetID", testSetID), zap.String("testRunID", testRunID), zap.Int("appID", appID))
	go func(testSetID, testRunID string, appID int) {
		ctx := context.WithoutCancel(ctx)
		status, err := r.replay.RunTestSet(ctx, testSetID, testRunID, uint64(appID), "", true)
		if err != nil {
			return
		}
		r.logger.Info("test set status", zap.String("status", string(status)))
	}(testSetID, testRunID, appID)

	return true, nil
}

// StartApp is the resolver for the startApp field.
func (r *mutationResolver) StartApp(ctx context.Context, appID int) (bool, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(utils.Emoji + "failed to get Resolver")
		return false, err
	}

	r.logger.Debug("starting application", zap.Int("appID", appID))

	appErrGrp, _ := errgroup.WithContext(ctx)
	appCtx := context.WithoutCancel(ctx)
	appCtx, appCancel := context.WithCancel(appCtx)
	appCtx = context.WithValue(appCtx, models.ErrGroupKey, appErrGrp)
	r.appCtx = appCtx
	r.appCancel = appCancel

	appErrGrp.Go(func() error {
		err := r.replay.RunApplication(appCtx, uint64(appID), models.RunOptions{})
		if err.Err != nil {
			r.logger.Error("failed to run application", zap.Error(err))
			utils.LogError(r.logger, err.Err, "error while running the application")
			return err
		}
		return nil
	})

	return true, nil
}

// TestSetStatus is the resolver for the testSetStatus field.
func (r *queryResolver) TestSetStatus(ctx context.Context, testRunID string, testSetID string) (*model.TestSetStatus, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(utils.Emoji + "failed to get Resolver")
		return nil, err
	}

	r.logger.Debug("getting test set status for", zap.String("testRunID", testRunID), zap.String("testSetID", testSetID))
	ctx = context.WithoutCancel(ctx)
	status, err := r.replay.GetTestSetStatus(ctx, testRunID, testSetID)
	if err != nil {
		utils.LogError(r.logger, err, "failed to get test set status")
		return nil, errors.New("failed to get test set status")
	}
	r.logger.Debug("test set status", zap.String("status", string(status)))
	return &model.TestSetStatus{
		Status: string(status),
	}, nil
}

// StopApp is the resolver for the stopApp field.
func (r *mutationResolver) StopApp(_ context.Context, appId int) (bool, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(utils.Emoji + "failed to get Resolver")
		return false, err
	}

	r.logger.Debug("stopping the application", zap.Int("appID", appId))
	appCtx := r.appCtx
	appCancel := r.appCancel

	if appCtx == nil {
		return false, fmt.Errorf("failed to get the app context")
	}
	g, ok := appCtx.Value(models.ErrGroupKey).(*errgroup.Group)
	if !ok {
		utils.LogError(r.logger, nil, "failed to get the app error group from the context")
		return false, errors.New("failed to stop the app")
	}

	// cancel the context of the app to stop the app
	if appCancel != nil {
		appCancel()
	}

	err := g.Wait()
	if err != nil {
		utils.LogError(r.logger, err, "failed to stop the app")
		return false, err
	}
	r.logger.Info("application stopped successfully", zap.Int("appID", appId))

	return true, nil
}

// StopHooks is the resolver for the stopHooks field.
func (r *mutationResolver) StopHooks(context.Context) (bool, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(utils.Emoji + "failed to get Resolver")
		return false, err
	}
	r.logger.Debug("stopping the hooks")
	err := utils.Stop(r.logger, "stopping the test run")
	if err != nil {
		utils.LogError(r.logger, err, "failed to stop the test run")
		return false, err
	}
	return true, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
