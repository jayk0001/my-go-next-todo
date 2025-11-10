package todo

import "context"

// ValidatorService handles todo input validation
type ValidatorService struct{}

// NewValidatorService creates a new validator service
func NewValidatorService() *ValidatorService {
	return &ValidatorService{}
}

// ValidateTodoInput validates create todo input
func (v *ValidatorService) ValidateTodoInput(ctx context.Context, input CreateTodoInput) error {
	if err := v.validateTitle(input.Title); err != nil {
		return err
	}

	if input.Description != nil {
		if err := v.validateDescription(*input.Description); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateInput validates update todo input
func (v *ValidatorService) ValidateUpdateInput(ctx context.Context, input UpdateTodoInput) error {
	// At least one field should be provided for update
	if input.Completed == nil && input.Description == nil && input.Title == nil {
		return ErrInvalidTodoInput
	}

	if input.Title != nil {
		if err := v.validateTitle(*input.Title); err != nil {
			return err
		}
	}

	if input.Description != nil {
		if err := v.validateDescription(*input.Description); err != nil {
			return err
		}
	}

	return nil
}

// ValidateTitle validates todo title
func (v *ValidatorService) validateTitle(title string) error {
	if title == "" {
		return ErrTodoTitleRequired
	}

	if len(title) > 500 {
		return ErrTodoTitleTooLong
	}

	return nil
}

// ValidateDescription validates todo description
func (v *ValidatorService) validateDescription(description string) error {
	if len(description) > 2000 {
		return ErrTodoDescriptionTooLong
	}

	return nil
}
