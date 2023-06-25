package approve

import (
	"time"

	"github.com/JIeeiroSst/workflow-service/internal/activities/approve/facade"
	"github.com/mitchellh/mapstructure"
	"go.temporal.io/sdk/workflow"
)

type ApproveWorkflow struct {
	facade facade.Facade
}

func NewApproveWorkflow(facade facade.Facade) *ApproveWorkflow {
	return &ApproveWorkflow{
		facade: facade,
	}
}

func (a *ApproveWorkflow) ApproveWorkflow(ctx workflow.Context, state ProcessState) error {
	err := workflow.SetQueryHandler(ctx, "getApprove", func(input []byte) (ProcessState, error) {
		return state, nil
	})
	if err != nil {
		return err
	}
	uploadChannel := workflow.GetSignalChannel(ctx, SignalChannels.UPLOAD_CHANNEL)
	processChannel := workflow.GetSignalChannel(ctx, SignalChannels.PROCESS_CHANNEL)
	approveChannel := workflow.GetSignalChannel(ctx, SignalChannels.APPROVE_CHANNEL)
	checkedOut := false
	sentProcess := false

	for {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(uploadChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
			var message UploadChannelSignal
			err := mapstructure.Decode(signal, &message)
			if err != nil {
				return
			}
			state.UploadApprove(message.Upload)
		})

		selector.AddReceive(processChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)

			var message ProcessChannelSignal
			err := mapstructure.Decode(signal, &message)
			if err != nil {
				return
			}
			state.ProcessApprove(message.Process)

			sentProcess = message.Process.IsApprove
		})

		selector.AddReceive(approveChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)

			var message ApproveChannelSignal
			err := mapstructure.Decode(signal, &message)
			if err != nil {
				return
			}

			ao := workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			}

			ctx := workflow.WithActivityOptions(ctx, ao)
			err = workflow.ExecuteActivity(ctx, state.ApproveProcess, state).Get(ctx, nil)
			if err != nil {
				return
			}

			checkedOut = true
		})

		if !sentProcess {
			selector.AddFuture(workflow.NewTimer(ctx, abandonedProcessTimeout), func(f workflow.Future) {
				ao := workflow.ActivityOptions{
					StartToCloseTimeout: time.Minute,
				}
				ctx = workflow.WithActivityOptions(ctx, ao)

				err := workflow.ExecuteActivity(ctx, state.SendAbandonedProcess, state).Get(ctx, nil)
				if err != nil {
					return
				}

			})
		}

		selector.Select(ctx)

		if checkedOut || sentProcess {
			break
		}
	}

	return nil
}
