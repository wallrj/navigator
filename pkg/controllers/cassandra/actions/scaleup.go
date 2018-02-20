package actions

import "github.com/jetstack/navigator/pkg/controllers"

type ScaleUp struct {
	Cluster   string
	Namespace string
	Replicas  int64
	NodePool  string
}

var _ controllers.Action = &ScaleUp{}

func (a *ScaleUp) Name() string {
	return "scaleup"
}

func (a *ScaleUp) Execute(s *controllers.State) error {
	return nil
}
