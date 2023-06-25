package approve

import "context"

type Activities struct{}

func (a *Activities) ApproveProcess(_ context.Context, process ProcessState) error {

	return nil
}

func (a *Activities) SendAbandonedProcess(_ context.Context, isApprove bool) error {

	return nil
}
