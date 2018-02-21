package actions

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
)

type ScaleOut struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &ScaleOut{}

func (a *ScaleOut) Name() string {
	return "ScaleOut"
}

func (a *ScaleOut) Execute(s *controllers.State) error {
	return nil
}
