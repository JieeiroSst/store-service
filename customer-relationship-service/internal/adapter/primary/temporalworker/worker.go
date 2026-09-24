package temporalworker

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	tc "github.com/JIeeiroSst/customer-relationship-service/internal/infrastructure/temporal"
	"github.com/JIeeiroSst/customer-relationship-service/internal/workflows"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

type Activities struct{ uc port.ContractLifecycleUsecase }

func NewActivities(uc port.ContractLifecycleUsecase) *Activities { return &Activities{uc: uc} }

func (a *Activities) Load(ctx context.Context, id uint) (port.ContractLifecycleState, error) {
	return a.uc.State(ctx, id)
}

func (a *Activities) Expire(ctx context.Context, in workflows.ExpireInput) (bool, error) {
	return a.uc.Expire(ctx, in.ContractID, in.At)
}

func (a *Activities) Remind(ctx context.Context, in workflows.RemindInput) error {
	return a.uc.Remind(ctx, in.ContractID, in.Remaining)
}

func Register(lc fx.Lifecycle, c *tc.Client, acts *Activities, uc port.ContractLifecycleUsecase) {
	if !c.Enabled() {
		return
	}
	w := worker.New(c.Client, c.Config.TaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.ContractLifecycle)
	w.RegisterActivityWithOptions(acts.Load, activity.RegisterOptions{Name: workflows.ActivityLoad})
	w.RegisterActivityWithOptions(acts.Expire, activity.RegisterOptions{Name: workflows.ActivityExpire})
	w.RegisterActivityWithOptions(acts.Remind, activity.RegisterOptions{Name: workflows.ActivityRemind})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := w.Start(); err != nil {
				return err
			}
			logrus.Infof("temporal worker started (queue %s, namespace %s)", c.Config.TaskQueue, c.Config.Namespace)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				n, err := uc.Resync(ctx)
				if err != nil {
					logrus.WithError(err).Error("contract lifecycle resync failed")
					return
				}
				logrus.WithField("contracts", n).Info("contract lifecycles synced")
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			w.Stop()
			return nil
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewActivities),
	fx.Invoke(Register),
)
