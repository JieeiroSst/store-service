package approve

import "go.temporal.io/sdk/workflow"

func ApproveWorkflow(ctx workflow.Context, state ProcessState) error {
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

	for {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(uploadChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
		})

		selector.AddReceive(processChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
		})

		selector.AddReceive(approveChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
		})

		selector.Select(ctx)

		if checkedOut {
			break
		}
	}

	return nil
}
