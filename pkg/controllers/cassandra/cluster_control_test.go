package cassandra_test

import (
	"reflect"
	"testing"
	"testing/quick"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/actions"
	"github.com/kr/pretty"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNextAction(t *testing.T) {
	cases := map[string]struct {
		c *v1alpha1.CassandraCluster
		a controllers.Action
	}{
		"scale up": {
			c: &v1alpha1.CassandraCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
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
			a: &actions.ScaleUp{
				Namespace: "foo",
				Cluster:   "bar",
				NodePool:  "np1",
				Replicas:  2,
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
					t.Errorf("Expected did not equal actual: %s", pretty.Diff(test.a, a))
				}
			},
		)
	}
}

func NodePoolsWithoutStatus(c v1alpha1.CassandraCluster) []string {
	nodepoolsWithoutStatus := []string{}
	for _, np := range c.Spec.NodePools {
		_, found := c.Status.NodePools[np.Name]
		if !found {
			nodepoolsWithoutStatus = append(nodepoolsWithoutStatus, np.Name)
		}
	}
	return nodepoolsWithoutStatus
}

func NodePoolsAllHaveStatus(c v1alpha1.CassandraCluster) bool {
	return len(NodePoolsWithoutStatus(c)) == 0
}

func TestQuick(t *testing.T) {
	f := func(c v1alpha1.CassandraCluster) bool {
		a, err := cassandra.NextAction(&c)
		// NextAction should never return an error
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return false
		}

		switch action := a.(type) {
		case *actions.CreateNodePool:
			_, found := c.Status.NodePools[action.NodePool]
			if found {
				t.Errorf("Unexpected attempt to create a nodepool when there's an existing status")
				return false
			}
		case *actions.ScaleUp:
			nps, found := c.Status.NodePools[action.NodePool]
			if !found {
				t.Errorf("Unexpected attempt to scale up a nodepool without a status")
				return false
			}
			if action.Replicas <= nps.ReadyReplicas {
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
