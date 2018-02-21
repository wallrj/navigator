package cassandra_test

import (
	"testing"
	"testing/quick"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/actions"
)

// func TestNextAction(t *testing.T) {
//	cases := map[string]struct {
//		c *v1alpha1.CassandraCluster
//		a controllers.Action
//	}{
//		"scale up": {
//			c: &v1alpha1.CassandraCluster{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "bar",
//					Namespace: "foo",
//				},
//				Spec: v1alpha1.CassandraClusterSpec{
//					NodePools: []v1alpha1.CassandraClusterNodePool{
//						{
//							Name:     "np1",
//							Replicas: 2,
//						},
//					},
//				},
//				Status: v1alpha1.CassandraClusterStatus{
//					NodePools: map[string]v1alpha1.CassandraClusterNodePoolStatus{
//						"np1": {
//							ReadyReplicas: 1,
//						},
//					},
//				},
//			},
//			a: &actions.ScaleOut{
//				Namespace: "foo",
//				Cluster:   "bar",
//				NodePool:  "np1",
//				Replicas:  2,
//			},
//		},
//	}

//	for title, test := range cases {
//		t.Run(
//			title,
//			func(t *testing.T) {
//				a := cassandra.NextAction(test.c)
//				if !reflect.DeepEqual(test.a, a) {
//					t.Errorf("Expected did not equal actual: %s", pretty.Diff(test.a, a))
//				}
//			},
//		)
//	}
// }

func TestQuick(t *testing.T) {
	f := func(c v1alpha1.CassandraCluster) bool {
		a := cassandra.NextAction(&c)

		switch action := a.(type) {
		case *actions.CreateNodePool:
			_, found := c.Status.NodePools[action.NodePool.Name]
			if found {
				t.Errorf("Unexpected attempt to create a nodepool when there's an existing status")
				return false
			}
		case *actions.ScaleOut:
			nps, found := c.Status.NodePools[action.NodePool.Name]
			if !found {
				t.Errorf("Unexpected attempt to scale up a nodepool without a status")
				return false
			}
			if action.NodePool.Replicas <= nps.ReadyReplicas {
				t.Errorf("Unexpected attempt to scale up a nodepool with >= ready replicas")
				return false
			}
		}
		return true
	}
	config := &quick.Config{
		MaxCount: 1000,
	}
	err := quick.Check(f, config)
	if err != nil {
		t.Errorf("quick check failure: %#v", err)
	}
}
