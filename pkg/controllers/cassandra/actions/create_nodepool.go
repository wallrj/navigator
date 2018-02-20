package actions

import "github.com/jetstack/navigator/pkg/controllers"

type CreateNodePool struct {
	Cluster   string
	Namespace string
	NodePool  string
}

var _ controllers.Action = &CreateNodePool{}

func (a *CreateNodePool) Name() string {
	return "CreateNodePool"
}

func (a *CreateNodePool) Execute(s *controllers.State) error {
	return nil
}
