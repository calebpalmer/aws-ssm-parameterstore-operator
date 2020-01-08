package controller

import (
	"github.com/calebpalmer/aws-ssm-secret-operator/pkg/controller/ssmparameter"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, ssmparameter.Add)
}
