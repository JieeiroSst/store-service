package internal

import validation "github.com/go-ozzo/ozzo-validation/v4"

type CreateParams struct {
	Description string
	Priority    Priority
	Dates       Dates
}

func (c CreateParams) Validate() error {
	if c.Priority == PriorityNone {
		return validation.Errors{
			"priority": NewErrorf(ErrorCodeInvalidArgument, "must be set"),
		}
	}

	task := Task{
		Description: c.Description,
		Priority:    c.Priority,
		Dates:       c.Dates,
	}

	if err := validation.Validate(&task); err != nil {
		return WrapErrorf(err, ErrorCodeInvalidArgument, "validation.Validate")
	}

	return nil
}

type SearchParams struct {
	Description *string
	Priority    *Priority
	IsDone      *bool
	From        int64
	Size        int64
}

func (a SearchParams) IsZero() bool {
	return a.Description == nil &&
		a.Priority == nil &&
		a.IsDone == nil
}

type SearchResults struct {
	Tasks []Task
	Total int64
}
