package errors_test

import (
	"fmt"

	"github.com/tzvetkoff-go/errors"
)

// NewController ...
func NewController() error {
	err := NewService()
	return errors.Propagate(err, "could not create service")
}

// NewService ...
func NewService() error {
	err := NewRepository()
	return errors.Propagate(err, "could not create repository")
}

// NewRepository ...
func NewRepository() error {
	return errors.New("could not connect to database")
}

func Example() {
	err := NewController()

	fmt.Println()
	fmt.Println(err.Error())

	fmt.Println()
	fmt.Println(errors.Cause(err))

	// Output:
	//
	// could not create service
	//  --- at example_test.go:12 (NewController) ---
	// Caused by: could not create repository
	//  --- at example_test.go:18 (NewService) ---
	// Caused by: could not connect to database
	//  --- at example_test.go:23 (NewRepository) ---
	//
	// could not connect to database
	//  --- at example_test.go:23 (NewRepository) ---
}
