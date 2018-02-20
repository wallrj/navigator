package cassandra

import (
	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/role"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/rolebinding"
	servicecql "github.com/jetstack/navigator/pkg/controllers/cassandra/service/cql"
	serviceseedprovider "github.com/jetstack/navigator/pkg/controllers/cassandra/service/seedprovider"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/serviceaccount"
)

const (
	ErrorSync = "ErrSync"

	SuccessSync = "SuccessSync"

	MessageErrorSyncServiceAccount = "Error syncing service account: %s"
	MessageErrorSyncRole           = "Error syncing role: %s"
	MessageErrorSyncRoleBinding    = "Error syncing role binding: %s"
	MessageErrorSyncConfigMap      = "Error syncing config map: %s"
	MessageErrorSyncService        = "Error syncing service: %s"
	MessageErrorSyncNodePools      = "Error syncing node pools: %s"
	MessageErrorSyncPilots         = "Error syncing pilots: %s"
	MessageSuccessSync             = "Successfully synced CassandraCluster"
)

type ControlInterface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

var _ ControlInterface = &defaultCassandraClusterControl{}

type defaultCassandraClusterControl struct {
	seedProviderServiceControl serviceseedprovider.Interface
	cqlServiceControl          servicecql.Interface
	nodepoolControl            nodepool.Interface
	pilotControl               pilot.Interface
	serviceAccountControl      serviceaccount.Interface
	roleControl                role.Interface
	roleBindingControl         rolebinding.Interface
	recorder                   record.EventRecorder
}

func NewControl(
	seedProviderServiceControl serviceseedprovider.Interface,
	cqlServiceControl servicecql.Interface,
	nodepoolControl nodepool.Interface,
	pilotControl pilot.Interface,
	serviceAccountControl serviceaccount.Interface,
	roleControl role.Interface,
	roleBindingControl rolebinding.Interface,
	recorder record.EventRecorder,
) ControlInterface {
	return &defaultCassandraClusterControl{
		seedProviderServiceControl: seedProviderServiceControl,
		cqlServiceControl:          cqlServiceControl,
		nodepoolControl:            nodepoolControl,
		pilotControl:               pilotControl,
		serviceAccountControl:      serviceAccountControl,
		roleControl:                roleControl,
		roleBindingControl:         roleBindingControl,
		recorder:                   recorder,
	}
}

func (e *defaultCassandraClusterControl) Sync(c *v1alpha1.CassandraCluster) error {
	glog.V(4).Infof("defaultCassandraClusterControl.Sync")
	err := e.seedProviderServiceControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncService,
			err,
		)
		return err
	}
	err = e.cqlServiceControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncService,
			err,
		)
		return err
	}
	err = e.nodepoolControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncNodePools,
			err,
		)
		return err
	}
	err = e.pilotControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncPilots,
			err,
		)
		return err
	}
	err = e.serviceAccountControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncServiceAccount,
			err,
		)
		return err
	}
	err = e.roleControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncRole,
			err,
		)
		return err
	}
	err = e.roleBindingControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncRoleBinding,
			err,
		)
		return err
	}

	e.recorder.Event(
		c,
		apiv1.EventTypeNormal,
		SuccessSync,
		MessageSuccessSync,
	)
	return nil
}

func NextAction(c *v1alpha1.CassandraCluster) (controllers.Action, error) {
	for _, np := range c.Spec.NodePools {
		nps, ok := c.Status.NodePools[np.Name]
		if !ok {
			continue
		}
		if np.Replicas > nps.ReadyReplicas {
			return &ScaleUp{
				Replicas: np.Replicas,
				NodePool: np.Name,
			}, nil
		}
	}
	return nil, nil
}

type ScaleUp struct {
	Replicas int64
	NodePool string
}

var _ controllers.Action = &ScaleUp{}

func (a *ScaleUp) Name() string {
	return "scaleup"
}

func (a *ScaleUp) Execute(s *controllers.State) error {
	return nil
}
