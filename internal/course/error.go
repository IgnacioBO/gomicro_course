package course

import (
	"errors"
	"fmt"
)

var ErrNameRequired = errors.New("name is required")
var ErrStartDateRequired = errors.New("start_date is required")
var ErrEndDateRequired = errors.New("end_date is required")

var ErrNameNotEmpty = errors.New("name can't be empty")
var ErrStartDateNotEmpty = errors.New("start_date can't be empty")
var ErrEndDateNotEmpty = errors.New("end_date can't be empty")

type ErrCourseNotFound struct {
	CourseID string
}

func (e ErrCourseNotFound) Error() string {
	return fmt.Sprintf("course with id: %s not found", e.CourseID)
}

type ErrDateBadFormat struct {
	DateName     string
	ErrorDetails string
}

func (e ErrDateBadFormat) Error() string {
	return fmt.Sprintf("invalid %s format: %s", e.DateName, e.ErrorDetails)
}
