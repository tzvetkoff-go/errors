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
	fmt.Println(err.Error())

	// Output:
	// could not create service
	//  --- at /home/w00t/Code/Go/errors/errors_test.go:12 (NewController) ---
	// Caused by: could not create repository
	//  --- at /home/w00t/Code/Go/errors/errors_test.go:18 (NewService) ---
	// Caused by: could not connect to database
	//  --- at /home/w00t/Code/Go/errors/errors_test.go:23 (NewRepository) ---
}
