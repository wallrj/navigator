package v1alpha1

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing/quick"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (_ CassandraCluster) Generate(rand *rand.Rand, size int) reflect.Value {
	o := CassandraCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("cluster%d", rand.Intn(10)),
			Namespace: "",
		},
	}
	v, ok := quick.Value(reflect.TypeOf(CassandraClusterSpec{}), rand)
	if ok {
		o.Spec = v.Interface().(CassandraClusterSpec)
	}
	v, ok = quick.Value(reflect.TypeOf(CassandraClusterStatus{}), rand)
	if ok {
		o.Status = v.Interface().(CassandraClusterStatus)
	}
	return reflect.ValueOf(o)
}

func (_ CassandraClusterSpec) Generate(rand *rand.Rand, size int) reflect.Value {
	nodepools := make([]CassandraClusterNodePool, rand.Intn(10))
	for i := range nodepools {
		v, ok := quick.Value(reflect.TypeOf(CassandraClusterNodePool{}), rand)
		if ok {
			nodepools[i] = v.Interface().(CassandraClusterNodePool)
		}
	}
	o := CassandraClusterSpec{
		CqlPort:   rand.Int31n(10),
		NodePools: nodepools,
	}
	return reflect.ValueOf(o)
}

func (_ CassandraClusterNodePool) Generate(rand *rand.Rand, size int) reflect.Value {
	o := CassandraClusterNodePool{
		Name:     fmt.Sprintf("np%d", rand.Intn(10)),
		Replicas: rand.Int31n(10),
	}
	return reflect.ValueOf(o)
}

func (_ CassandraClusterStatus) Generate(rand *rand.Rand, size int) reflect.Value {
	o := CassandraClusterStatus{
		NodePools: map[string]CassandraClusterNodePoolStatus{},
	}
	nodepools := make([]CassandraClusterNodePool, rand.Intn(10))
	for i := range nodepools {
		v, ok := quick.Value(reflect.TypeOf(CassandraClusterNodePool{}), rand)
		if ok {
			nodepools[i] = v.Interface().(CassandraClusterNodePool)
		}
	}
	for _, np := range nodepools {
		o.NodePools[np.Name] = CassandraClusterNodePoolStatus{
			ReadyReplicas: np.Replicas,
		}
	}
	return reflect.ValueOf(o)
}
