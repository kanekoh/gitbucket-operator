package controller

import (
	"github.com/kanekoh/gitbucket-operator/pkg/controller/gitbucket"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, gitbucket.Add)
}
