package course

import (
	"errors"
	"fmt"
)

var ErrNameRequired = errors.New("name is required")
var ErrStartDateRequired = errors.New("start_date is required")
var ErrEndDateRequired = errors.New("end_date is required")

type ErrCourseNotFound struct {
	CourseID string
}

func (e ErrCourseNotFound) Error() string {
	return fmt.Sprintf("course with id: %s not found", e.CourseID)
}
