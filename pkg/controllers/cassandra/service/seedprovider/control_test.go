package seedprovider_test

import (
	"fmt"
	"testing"

	apiv1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/service/seedprovider"
	servicetesting "github.com/jetstack/navigator/pkg/controllers/cassandra/service/testing"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func newService(f *casstesting.Fixture) *apiv1.Service {
	return seedprovider.ServiceForCluster(f.Cluster)
}

func TestSeedProviderServiceSync(t *testing.T) {
	servicetesting.RunStandardServiceTests(t, casstesting.NewFixture, newService)
	t.Run(
		"sync error",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.SeedProviderServiceControl = &casstesting.FakeControl{
				SyncError: fmt.Errorf("simulated sync error"),
			}
			f.RunExpectError()
		},
	)
}
