package cassandra_test

import (
	"reflect"
	"testing"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
)

func TestNextAction(t *testing.T) {
	cases := map[string]struct {
		c *v1alpha1.CassandraCluster
		a controllers.Action
	}{
		"scale up": {
			c: &v1alpha1.CassandraCluster{
				Spec: v1alpha1.CassandraClusterSpec{
					NodePools: []v1alpha1.CassandraClusterNodePool{
						{
							Name:     "np1",
							Replicas: 2,
						},
					},
				},
				Status: v1alpha1.CassandraClusterStatus{
					NodePools: map[string]v1alpha1.CassandraClusterNodePoolStatus{
						"np1": {
							ReadyReplicas: 1,
						},
					},
				},
			},
			a: &cassandra.ScaleUp{
				Replicas: 2,
				NodePool: "np1",
			},
		},
	}

	for title, test := range cases {
		t.Run(
			title,
			func(t *testing.T) {
				a, err := cassandra.NextAction(test.c)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(test.a, a) {
					t.Errorf("Expected %#v. Got %#v", test.a, a)
				}
			},
		)
	}
}
