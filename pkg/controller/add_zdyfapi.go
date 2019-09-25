package controller

import (
	"zdyfop/pkg/controller/zdyfapi"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, zdyfapi.Add)
}
