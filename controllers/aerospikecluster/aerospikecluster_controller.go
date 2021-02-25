package aerospikecluster

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	as "github.com/ashishshinde/aerospike-client-go"

	accessControl "github.com/aerospike/aerospike-kubernetes-operator/controllers/asconfig"
	"github.com/aerospike/aerospike-kubernetes-operator/controllers/configmap"
	"github.com/aerospike/aerospike-kubernetes-operator/controllers/jsonpatch"

	"github.com/aerospike/aerospike-kubernetes-operator/controllers/utils"
	lib "github.com/aerospike/aerospike-management-lib"
	log "github.com/inconshreveable/log15"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	aerospikev1alpha1 "github.com/aerospike/aerospike-kubernetes-operator/api/v1alpha1"
)

const defaultUser = "admin"
const defaultPass = "admin"
const patchFieldOwner = "aerospike-kuberneter-operator"
const finalizerName = "aerospike.aerospike.com/storage-finalizer"

// Number of reconcile threads to run reconcile operations
var maxConcurrentReconciles = runtime.NumCPU() * 2

var (
	updateOption = &client.UpdateOptions{
		FieldManager: "aerospike-operator",
	}
	createOption = &client.CreateOptions{
		FieldManager: "aerospike-operator",
	}
)

// func ignoreDeletionPredicate() predicate.Predicate {
// 	return predicate.Funcs{
// 		UpdateFunc: func(e event.UpdateEvent) bool {
// 			// Ignore updates to CR status in which case metadata.Generation does not change
// 			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
// 		},
// 		DeleteFunc: func(e event.DeleteEvent) bool {
// 			// Evaluates to false if the object has been confirmed deleted.
// 			return !e.DeleteStateUnknown
// 		},
// 	}
// }

func ignoreSecondaryResource() predicate.Predicate {
	// Watch for changes to secondary resource StatefulSet and requeue the owner AerospikeCluster
	// err = c.Watch(
	// 	&source.Kind{Type: &appsv1.StatefulSet{}},
	// 	&handler.EnqueueRequestForOwner{
	// 		IsController: true,
	// 		OwnerType:    &aerospikev1alpha1.AerospikeCluster{},
	// 	}, predicate.Funcs{
	// 		CreateFunc: func(e event.CreateEvent) bool {
	// 			return false
	// 		},
	// 		UpdateFunc: func(e event.UpdateEvent) bool {
	// 			return false
	// 		},
	// 	)
	// if err != nil {
	// 	return err
	// }
	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
	return p
}

// SetupWithManager sets up the controller with the Manager
func (r *AerospikeClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aerospikev1alpha1.AerospikeCluster{}).
		// Watches(
		// 	&source.Kind{Type: &appsv1.StatefulSet{}},
		// 	&handler.EnqueueRequestForOwner{
		// 		IsController: true,
		// 		OwnerType:    &aerospikev1alpha1.AerospikeCluster{},
		// 	},

		// 	builder.WithPredicates(
		// 		predicate.Funcs{
		// 			CreateFunc: func(e event.CreateEvent) bool {
		// 				return false
		// 			},
		// 			UpdateFunc: func(e event.UpdateEvent) bool {
		// 				return false
		// 			},
		// 		},
		// 	),
		// ).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(
			predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					return false
				},
			},
		)).
		// WithOptions(controller.Options{
		// 	MaxConcurrentReconciles: maxConcurrentReconciles,
		// }).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		// WithEventFilter(predicate.Funcs{
		// 	UpdateFunc: func(e event.UpdateEvent) bool {
		// 		// Generation is only updated on spec changes (also on deletion),
		// 		// not metadata or status
		// 		// Filter out events where the generation hasn't changed to
		// 		// avoid being triggered by status updates
		// 		return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		// 	},
		// 	DeleteFunc: func(e event.DeleteEvent) bool {
		// 		// The reconciler adds a finalizer so we perform clean-up
		// 		// when the delete timestamp is added
		// 		// Suppress Delete events to avoid filtering them out in the Reconcile function
		// 		return false
		// 	},
		// }).
		Complete(r)
}

// // Add creates a new AerospikeCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// // and Start it when the Manager is Started.
// func Add(mgr manager.Manager) error {
// 	return add(mgr, newReconciler(mgr))
// }

// // newReconciler returns a new reconcile.Reconciler
// func newReconciler(mgr manager.Manager) reconcile.Reconciler {
// 	return &AerospikeClusterReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
// }

// // add adds a new Controller to mgr with r as the reconcile.Reconciler
// func add(mgr manager.Manager, r reconcile.Reconciler) error {
// 	// Create a new controller
// 	c, err := controller.New("aerospikecluster-controller", mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: maxConcurrentReconciles)
// 	if err != nil {
// 		return err
// 	}

// 	// Watch for changes to primary resource AerospikeCluster
// 	err = c.Watch(
// 		&source.Kind{Type: &aerospikev1alpha1.AerospikeCluster{}},
// 		&handler.EnqueueRequestForObject{},
// 		// Skip where cluster object generation is not changed
// 		predicate.GenerationChangedPredicate{)
// 	if err != nil {
// 		return err
// 	}

// 	// TODO: Do we need to monitor this? Statefulset is updated many times in reconcile and this add new entry in
// 	// update queue. If user will change only cr then we may not need to monitor statefulset.
// 	// Think all possible situation

// 	// Watch for changes to secondary resource StatefulSet and requeue the owner AerospikeCluster
// 	err = c.Watch(
// 		&source.Kind{Type: &appsv1.StatefulSet{}},
// 		&handler.EnqueueRequestForOwner{
// 			IsController: true,
// 			OwnerType:    &aerospikev1alpha1.AerospikeCluster{},
// 		}, predicate.Funcs{
// 			CreateFunc: func(e event.CreateEvent) bool {
// 				return false
// 			},
// 			UpdateFunc: func(e event.UpdateEvent) bool {
// 				return false
// 			},
// 		)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

var pkglog = log.New("module", "controller_aerospikecluster")

// // blank assignment to verify that AerospikeClusterReconciler implements reconcile.Reconciler
// var _ reconcile.Reconciler = &AerospikeClusterReconciler{}

// AerospikeClusterReconciler reconciles a AerospikeCluster object
type AerospikeClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *k8sRuntime.Scheme
}

// RackState contains the rack configuration and rack size.
type RackState struct {
	Rack aerospikev1alpha1.Rack
	Size int
}

// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aerospike.aerospike.com,resources=aerospikeclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aerospike.aerospike.com,resources=aerospikeclusters/status,verbs=get;update;patch;

// Reconcile AerospikeCluster object
func (r *AerospikeClusterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// logger := pkglog.New("AerospikeCluster", request.NamespacedName)
	_ = r.Log.WithValues("aerospikecluster", request.NamespacedName)

	r.Log.Info("Reconciling AerospikeCluster")

	// Fetch the AerospikeCluster instance
	aeroCluster := &aerospikev1alpha1.AerospikeCluster{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, aeroCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{Requeue: true}, err
	}

	r.Log.V(1).Info("AerospikeCluster", "Spec", aeroCluster.Spec, "Status", aeroCluster.Status)

	// finalizerName := "storage.finalizer"

	// Check DeletionTimestamp to see if cluster is being deleted
	if !aeroCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// TODO: LOG FOR CLUSTER DELETION
		// The cluster is being deleted
		if err := r.handleClusterDeletion(aeroCluster, finalizerName); err != nil {
			return ctrl.Result{}, err
		}
		// Stop reconciliation as the cluster is being deleted
		return ctrl.Result{}, nil
	}

	// The cluster is not being deleted, add finalizer in not added already
	if err := r.addFinalizer(aeroCluster, finalizerName); err != nil {
		r.Log.Error(err, "Failed to add finalizer")
		return ctrl.Result{}, err
	}

	// Handle previously failed cluster
	if err := r.handlePreviouslyFailedCluster(aeroCluster); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile all racks
	if res := r.reconcileRacks(aeroCluster); !res.isSuccess {
		if res.err != nil {
			r.Log.Error(res.err, "Failed to reconcile racks")
		}
		return res.result, res.err
	}

	// Setup access control.
	if err := r.reconcileAccessControl(aeroCluster); err != nil {
		r.Log.Error(err, "Failed to reconcile access control")
		return ctrl.Result{}, err
	}

	// Update the AerospikeCluster status.
	if err := r.updateStatus(aeroCluster); err != nil {
		r.Log.Error(err, "Failed to update AerospikeCluster status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AerospikeClusterReconciler) handleClusterDeletion(aeroCluster *aerospikev1alpha1.AerospikeCluster, finalizerName string) error {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))
	r.Log.Info("Handle cluster deletion")

	// The cluster is being deleted
	if err := r.cleanUpAndremoveFinalizer(aeroCluster, finalizerName); err != nil {
		r.Log.Error(err, "Failed to remove finalizer")
		return err
	}
	return nil
}

func (r *AerospikeClusterReconciler) handlePreviouslyFailedCluster(aeroCluster *aerospikev1alpha1.AerospikeCluster) error {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))
	r.Log.Info("Handle previously failed cluster")

	isNew, err := r.isNewCluster(aeroCluster)
	if err != nil {
		return fmt.Errorf("Error determining if cluster is new: %v", err)
	}

	if isNew {
		r.Log.V(1).Info("It's new cluster, create empty status object")
		if err := r.createStatus(aeroCluster); err != nil {
			r.Log.Error(err, "Failed to create empty status")
			return err
		}
	} else {
		r.Log.V(1).Info("It's not a new cluster, check if it is failed and needs recovery")
		hasFailed, err := r.hasClusterFailed(aeroCluster)
		if err != nil {
			r.Log.Error(err, "Failed to determining if cluster has failed")
			return fmt.Errorf("Error determining if cluster has failed: %v", err)
		}

		if hasFailed {
			return r.recoverFailedCreate(aeroCluster)
		}
	}
	return nil
}

func (r *AerospikeClusterReconciler) getResourceVersion(stsName types.NamespacedName) string {
	found := &appsv1.StatefulSet{}
	r.Client.Get(context.TODO(), stsName, found)
	return found.ObjectMeta.ResourceVersion
}

func (r *AerospikeClusterReconciler) reconcileRacks(aeroCluster *aerospikev1alpha1.AerospikeCluster) reconcileResult {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))
	r.Log.Info("Reconciling rack for AerospikeCluster")

	var scaledDownRackSTSList []appsv1.StatefulSet
	var scaledDownRackList []RackState
	var res reconcileResult

	rackStateList := getNewRackStateList(aeroCluster)
	for _, state := range rackStateList {
		found := &appsv1.StatefulSet{}
		stsName := getNamespacedNameForStatefulSet(aeroCluster, state.Rack.ID)
		if err := r.Client.Get(context.TODO(), stsName, found); err != nil {
			if !errors.IsNotFound(err) {
				return reconcileError(err)
			}

			// Create statefulset with 0 size rack and then scaleUp later in reconcile
			zeroSizedRack := RackState{Rack: state.Rack, Size: 0}
			found, res = r.createRack(aeroCluster, zeroSizedRack)
			if !res.isSuccess {
				return res
			}
		}

		// Get list of scaled down racks
		if *found.Spec.Replicas > int32(state.Size) {
			scaledDownRackSTSList = append(scaledDownRackSTSList, *found)
			scaledDownRackList = append(scaledDownRackList, state)
		} else {
			// Reconcile other statefulset
			if res := r.reconcileRack(aeroCluster, found, state); !res.isSuccess {
				return res
			}
		}
	}

	// Reconcile scaledDownRacks after all other racks are reconciled
	for idx, state := range scaledDownRackList {
		if res := r.reconcileRack(aeroCluster, &scaledDownRackSTSList[idx], state); !res.isSuccess {
			return res
		}
	}

	if len(aeroCluster.Status.RackConfig.Racks) != 0 {
		// Remove removed racks
		if res := r.deleteRacks(aeroCluster, rackStateList); !res.isSuccess {
			r.Log.Error(res.err, "Failed to remove statefulset for removed racks")
			return res
		}
	}

	return reconcileSuccess()
}

func (r *AerospikeClusterReconciler) createRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, rackState RackState) (*appsv1.StatefulSet, reconcileResult) {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))
	r.Log.Info("Create new Aerospike cluster if needed")

	// NoOp if already exist
	r.Log.Info("AerospikeCluster", "Spec", aeroCluster.Spec)
	if err := r.createHeadlessSvc(aeroCluster); err != nil {
		r.Log.Error(err, "Failed to create headless service")
		return nil, reconcileError(err)
	}

	// Bad config should not come here. It should be validated in validation hook
	cmName := getNamespacedNameForConfigMap(aeroCluster, rackState.Rack.ID)
	if err := r.buildConfigMap(aeroCluster, cmName, rackState.Rack); err != nil {
		r.Log.Error(err, "Failed to create configMap from AerospikeConfig")
		return nil, reconcileError(err)
	}

	stsName := getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID)
	found, err := r.createStatefulSet(aeroCluster, stsName, rackState)
	if err != nil {
		r.Log.Error(err, "Statefulset setup failed. Deleting statefulset", "name", stsName, "err", err)
		// Delete statefulset and everything related so that it can be properly created and updated in next run
		r.deleteStatefulSet(aeroCluster, found)
		return nil, reconcileError(err)
	}
	return found, reconcileSuccess()
}

func (r *AerospikeClusterReconciler) deleteRacks(aeroCluster *aerospikev1alpha1.AerospikeCluster, rackStateList []RackState) reconcileResult {
	oldRackList := getOldRackList(aeroCluster)

	for _, rack := range oldRackList {
		var rackFound bool
		for _, newRack := range rackStateList {
			if rack.ID == newRack.Rack.ID {
				rackFound = true
				break
			}
		}

		if rackFound {
			continue
		}

		found := &appsv1.StatefulSet{}
		stsName := getNamespacedNameForStatefulSet(aeroCluster, rack.ID)
		err := r.Client.Get(context.TODO(), stsName, found)
		if err != nil {
			// If not found then go to next
			if errors.IsNotFound(err) {
				continue
			}
			return reconcileError(err)
		}
		// TODO: Add option for quick delete of rack. DefaultRackID should always be removed gracefully
		found, res := r.scaleDownRack(aeroCluster, found, RackState{Size: 0, Rack: rack})
		if !res.isSuccess {
			return res
		}

		// Delete sts
		if err := r.deleteStatefulSet(aeroCluster, found); err != nil {
			return reconcileError(err)
		}
	}
	return reconcileSuccess()
}

func (r *AerospikeClusterReconciler) reconcileRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, found *appsv1.StatefulSet, rackState RackState) reconcileResult {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))
	r.Log.Info("Reconcile existing Aerospike cluster statefulset")

	var err error
	var res reconcileResult

	r.Log.Info("Ensure rack StatefulSet size is the same as the spec")
	desiredSize := int32(rackState.Size)
	// Scale down
	if *found.Spec.Replicas > desiredSize {
		found, res = r.scaleDownRack(aeroCluster, found, rackState)
		if !res.isSuccess {
			if res.err != nil {
				r.Log.Error(err, "Failed to scaleDown StatefulSet pods")
			}
			return res
		}
	}

	// isRackAerospikeConfigUpdated := r.isAerospikeConfigUpdatedForRack(aeroCluster, rackState)
	// // Update config map if config is updated
	// if isRackAerospikeConfigUpdated {
	// 	if err := r.updateConfigMap(aeroCluster, getNamespacedNameForConfigMap(aeroCluster, rackState.Rack.ID), rackState); err != nil {
	// 		r.Log.Error(err, "Failed to update configMap from AerospikeConfig")
	// 		return reconcileError(err)
	// 	}
	// }

	// Always update configMap. We won't be able to find if a rack's config and it's pod config is in sync or not
	// Checking rack.spec, rack.status will not work.
	// We may change config, let some pods restart with new config and then change config back to original value.
	// Now rack.spec, rack.status will be same but few pods will have changed config.
	// So a check based on spec and status will skip configMap update.
	// Hence a rolling restart of pod will never bring pod to desired config
	if err := r.updateConfigMap(aeroCluster, getNamespacedNameForConfigMap(aeroCluster, rackState.Rack.ID), rackState.Rack); err != nil {
		r.Log.Error(err, "Failed to update configMap from AerospikeConfig")
		return reconcileError(err)
	}

	// Upgrade
	upgradeNeeded, err := r.isAeroClusterUpgradeNeeded(aeroCluster, rackState.Rack.ID)
	if err != nil {
		return reconcileError(err)
	}

	if upgradeNeeded {
		found, res = r.upgradeRack(aeroCluster, found, aeroCluster.Spec.Image, rackState)
		if !res.isSuccess {
			if res.err != nil {
				r.Log.Error(res.err, "Failed to update StatefulSet image")
			}
			return res
		}
	} else {
		needRollingRestartRack, err := r.needRollingRestartRack(aeroCluster, rackState)
		if err != nil {
			return reconcileError(err)
		}
		if needRollingRestartRack {
			found, res = r.rollingRestartRack(aeroCluster, found, rackState)
			if !res.isSuccess {
				if res.err != nil {
					r.Log.Error(res.err, "Failed to do rolling restart")
				}
				return res
			}
		}
	}

	// Scale up after upgrading, so that new pods comeup with new image
	if *found.Spec.Replicas < desiredSize {
		found, res = r.scaleUpRack(aeroCluster, found, rackState)
		if !res.isSuccess {
			r.Log.Error(res.err, "Failed to scaleUp StatefulSet pods")
			return res
		}
	}

	// TODO: check if all the pods are up or not
	return reconcileSuccess()
}

func isClusterAerospikeConfigSecretUpdated(aeroCluster *aerospikev1alpha1.AerospikeCluster) bool {
	return !reflect.DeepEqual(aeroCluster.Spec.AerospikeConfigSecret, aeroCluster.Status.AerospikeConfigSecret)
}

func (r *AerospikeClusterReconciler) needRollingRestartRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, rackState RackState) (bool, error) {
	podList, err := r.getOrderedRackPodList(aeroCluster, rackState.Rack.ID)
	if err != nil {
		return false, fmt.Errorf("Failed to list pods: %v", err)
	}
	for _, pod := range podList {
		// Check if this pod need restart
		needRollingRestart, err := r.needRollingRestartPod(aeroCluster, rackState, pod)
		if err != nil {
			return false, err
		}
		if needRollingRestart {
			return true, nil
		}
	}
	return false, nil
}

func (r *AerospikeClusterReconciler) isAerospikeConfigUpdatedForRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, rackState RackState) bool {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))

	// AerospikeConfig nil means status not updated yet
	if aeroCluster.Status.AerospikeConfig.Raw == nil {
		return false
	}
	// TODO: Should we use some other check? Reconcile may requeue request multiple times before updating status and
	// then this func will return true until status is updated.
	// No need to check global AerospikeConfig. Racks will always have
	// Only Check if rack AerospikeConfig has changed
	for _, statusRack := range aeroCluster.Status.RackConfig.Racks {
		if rackState.Rack.ID == statusRack.ID {
			if !reflect.DeepEqual(rackState.Rack.AerospikeConfig, statusRack.AerospikeConfig) {
				r.Log.Info("Rack AerospikeConfig changed. Need rolling restart", "oldRackConfig", statusRack, "newRackConfig", rackState.Rack)
				return true
			}

			break
		}
	}

	return false
}

func (r *AerospikeClusterReconciler) scaleUpRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, found *appsv1.StatefulSet, rackState RackState) (*appsv1.StatefulSet, reconcileResult) {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))

	desiredSize := int32(rackState.Size)

	oldSz := *found.Spec.Replicas
	found.Spec.Replicas = &desiredSize

	r.Log.Info("Scaling up pods", "currentSz", oldSz, "desiredSz", desiredSize)

	// No need for this? But if image is bad then new pod will also comeup with bad node.
	podList, err := r.getRackPodList(aeroCluster, rackState.Rack.ID)
	if err != nil {
		return found, reconcileError(fmt.Errorf("Failed to list pods: %v", err))
	}
	if r.isAnyPodInFailedState(aeroCluster, podList.Items) {
		return found, reconcileError(fmt.Errorf("Cannot scale up AerospikeCluster. A pod is already in failed state"))
	}

	newPodNames := []string{}
	for i := oldSz; i < desiredSize; i++ {
		newPodNames = append(newPodNames, getStatefulSetPodName(found.Name, i))
	}

	// Ensure none of the to be launched pods are active.
	for _, newPodName := range newPodNames {
		for _, pod := range podList.Items {
			if pod.Name == newPodName {
				return found, reconcileError(fmt.Errorf("Pod %s yet to be launched is still present", newPodName))
			}
		}
	}

	// Clean up any dangling resources associated with the new pods.
	// This implements a safety net to protect scale up against failed cleanup operations when cluster
	// is scaled down.
	cleanupPods := []string{}
	cleanupPods = append(cleanupPods, newPodNames...)

	// Find dangling pods in pods
	if aeroCluster.Status.Pods != nil {
		for podName := range aeroCluster.Status.Pods {
			ordinal, err := getStatefulSetPodOrdinal(podName)
			if err != nil {
				return found, reconcileError(fmt.Errorf("Invalid pod name: %s", podName))
			}

			if *ordinal >= desiredSize {
				cleanupPods = append(cleanupPods, podName)
			}
		}
	}

	err = r.cleanupPods(aeroCluster, cleanupPods, rackState)
	if err != nil {
		return found, reconcileError(fmt.Errorf("Failed scale up pre-check: %v", err))
	}

	if aeroCluster.Spec.MultiPodPerHost {
		// Create services for each pod
		for _, podName := range newPodNames {
			if err := r.createServiceForPod(aeroCluster, podName, aeroCluster.Namespace); err != nil {
				return found, reconcileError(err)
			}
		}
	}

	// Scale up the statefulset
	if err := r.Client.Update(context.TODO(), found, updateOption); err != nil {
		return found, reconcileError(fmt.Errorf("Failed to update StatefulSet pods: %v", err))
	}

	if err := r.waitForStatefulSetToBeReady(found); err != nil {
		return found, reconcileError(fmt.Errorf("Failed to wait for statefulset to be ready: %v", err))
	}

	// return a fresh copy
	found, err = r.getStatefulSet(aeroCluster, rackState)
	if err != nil {
		return found, reconcileError(err)
	}
	return found, reconcileSuccess()
}

func (r *AerospikeClusterReconciler) upgradeRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, found *appsv1.StatefulSet, desiredImage string, rackState RackState) (*appsv1.StatefulSet, reconcileResult) {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))

	// List the pods for this aeroCluster's statefulset
	podList, err := r.getOrderedRackPodList(aeroCluster, rackState.Rack.ID)
	if err != nil {
		return found, reconcileError(fmt.Errorf("Failed to list pods: %v", err))
	}

	// Update strategy for statefulSet is OnDelete, so client.Update will not start update.
	// Update will happen only when a pod is deleted.
	// So first update image and then delete a pod. Pod will come up with new image.
	// Repeat the above process.
	needsUpdate := false
	for i, container := range found.Spec.Template.Spec.Containers {
		desiredImage, err := utils.GetDesiredImage(aeroCluster, container.Name)

		if err != nil {
			// Maybe a deleted sidecar.
			continue
		}

		if !utils.IsImageEqual(container.Image, desiredImage) {
			r.Log.Info("Updating image in statefulset spec", "container", container.Name, "desiredImage", desiredImage, "currentImage", container.Image)
			found.Spec.Template.Spec.Containers[i].Image = desiredImage
			needsUpdate = true
		}
	}

	if needsUpdate {
		err = r.Client.Update(context.TODO(), found, updateOption)
		if err != nil {
			return found, reconcileError(fmt.Errorf("Failed to update image for StatefulSet %s: %v", found.Name, err))
		}
	}

	for _, p := range podList {
		r.Log.Info("Check if pod needs upgrade or not")
		var needPodUpgrade bool
		for _, ps := range p.Spec.Containers {
			desiredImage, err := utils.GetDesiredImage(aeroCluster, ps.Name)
			if err != nil {
				continue
			}

			if !utils.IsImageEqual(ps.Image, desiredImage) {
				r.Log.Info("Upgrading/downgrading pod", "podName", p.Name, "currentImage", ps.Image, "desiredImage", desiredImage)
				needPodUpgrade = true
				break
			}
		}
		if !needPodUpgrade {
			r.Log.Info("Pod doesn't need upgrade")
			continue
		}
		// Also check if statefulSet is in stable condition
		// Check for all containers. Status.ContainerStatuses doesn't include init container
		res := r.checkPodImageUpdated(aeroCluster, desiredImage, rackState, p)
		if !res.isSuccess {
			return found, res
		}
		return found, reconcileRequeueAfter(0)
	}

	// return a fresh copy
	found, err = r.getStatefulSet(aeroCluster, rackState)
	if err != nil {
		return found, reconcileError(err)
	}
	return found, reconcileSuccess()
}

func (r *AerospikeClusterReconciler) checkPodImageUpdated(aeroCluster *aerospikev1alpha1.AerospikeCluster, desiredImage string, rackState RackState, p corev1.Pod) reconcileResult {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))

	needsDeletion := false
	// Also check if statefulSet is in stable condition
	// Check for all containers. Spec.Containers doesn't include init container
	for _, ps := range p.Spec.Containers {
		desiredImage, err := utils.GetDesiredImage(aeroCluster, ps.Name)

		if err != nil {
			// Maybe a deleted sidecar.
			continue
		}

		if utils.IsImageEqual(ps.Image, desiredImage) {
			if err := utils.CheckPodFailed(&p); err != nil {
				// Looks like bad image
				return reconcileError(err)
			}
		}

		if !utils.IsImageEqual(ps.Image, desiredImage) {
			r.Log.Info("Upgrading/downgrading pod", "podName", p.Name, "currentImage", ps.Image, "desiredImage", desiredImage)
			needsDeletion = true
			break
		}
	}
	if needsDeletion {
		r.Log.V(1).Info("Delete the Pod", "podName", p.Name)

		// If already dead node, so no need to check node safety, migration
		if err := utils.CheckPodFailed(&p); err == nil {
			if res := r.waitForNodeSafeStopReady(aeroCluster, &p); !res.isSuccess {
				return res
			}
		}

		// Delete pod
		if err := r.Client.Delete(context.TODO(), &p); err != nil {
			return reconcileError(err)
		}
		r.Log.V(1).Info("Pod deleted", "podName", p.Name)

		// Wait for pod to come up
		for {
			r.Log.V(1).Info("Waiting for pod to be ready after delete", "podName", p.Name)

			pFound := &corev1.Pod{}
			err := r.Client.Get(context.TODO(), types.NamespacedName{Name: p.Name, Namespace: p.Namespace}, pFound)
			if err != nil {
				r.Log.Error(err, "Failed to get pod", "podName", p.Name, "err", err)

				if _, err = r.getStatefulSet(aeroCluster, rackState); err != nil {
					// Stateful set has been deleted.
					// TODO Ashish to rememeber which scenario this can happen.
					r.Log.Error(err, "Statefulset has been deleted for pod", "podName", p.Name, "err", err)
					return reconcileError(err)
				}

				time.Sleep(time.Second * 5)
				continue
			}
			if err := utils.CheckPodFailed(pFound); err != nil {
				return reconcileError(err)
			}
			if !utils.IsPodUpgraded(pFound, aeroCluster) {
				r.Log.V(1).Info("Waiting for pod to come up with new image", "podName", p.Name)
				time.Sleep(time.Second * 5)
				continue
			}

			r.Log.Info("Pod is upgraded/downgraded", "podName", p.Name)
			break
		}
	}

	return reconcileSuccess()
}

func (r *AerospikeClusterReconciler) rollingRestartRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, found *appsv1.StatefulSet, rackState RackState) (*appsv1.StatefulSet, reconcileResult) {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))

	r.Log.Info("Rolling restart AerospikeCluster statefulset nodes with new config")

	// List the pods for this aeroCluster's statefulset
	podList, err := r.getOrderedRackPodList(aeroCluster, rackState.Rack.ID)
	if err != nil {
		return found, reconcileError(fmt.Errorf("Failed to list pods: %v", err))
	}
	if r.isAnyPodInFailedState(aeroCluster, podList) {
		return found, reconcileError(fmt.Errorf("Cannot Rolling restart AerospikeCluster. A pod is already in failed state"))
	}

	// Can we optimize this? Update stateful set only if there is any update for it.
	//updateStatefulSetStorage(aeroCluster, found)

	updateStatefulSetPodSpec(aeroCluster, found)

	updateStatefulSetAerospikeServerContainerResources(aeroCluster, found)

	r.updateStatefulSetSecretInfo(aeroCluster, found)

	updateStatefulSetConfigMapVolumes(aeroCluster, found, rackState)

	r.Log.Info("Updating statefulset spec")

	if err := r.Client.Update(context.TODO(), found, updateOption); err != nil {
		return found, reconcileError(fmt.Errorf("Failed to update StatefulSet %s: %v", found.Name, err))
	}
	r.Log.Info("Statefulset spec updated. Doing rolling restart with new config")

	for _, pod := range podList {
		// Check if this pod need restart
		needRollingRestart, err := r.needRollingRestartPod(aeroCluster, rackState, pod)
		if err != nil {
			return found, reconcileError(err)
		}
		if !needRollingRestart {
			r.Log.Info("This Pod doesn't need rolling restart, Skip this")
			continue
		}

		res := r.rollingRestartPod(aeroCluster, rackState, pod)
		if !res.isSuccess {
			return found, res
		}
		return found, reconcileRequeueAfter(0)
	}

	// return a fresh copy
	found, err = r.getStatefulSet(aeroCluster, rackState)
	if err != nil {
		return found, reconcileError(err)
	}
	return found, reconcileSuccess()
}

func (r *AerospikeClusterReconciler) needRollingRestartPod(aeroCluster *aerospikev1alpha1.AerospikeCluster, rackState RackState, pod corev1.Pod) (bool, error) {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID), "Pod", pod.Name)

	needRollingRestartPod := false

	// AerospikeConfig nil means status not updated yet
	if aeroCluster.Status.AerospikeConfig.Raw == nil {
		return needRollingRestartPod, nil
	}

	// confStr, err := configmap.BuildConfigTemplate(aeroCluster, rackState.Rack)
	// if err != nil {
	// 	return false, err
	// }
	// confHash, err := utils.GetHash(confStr)
	// if err != nil {
	// 	return false, err
	// }

	// currentConfHash := pod.Annotations["aerospike-config-hash"]

	// if currentConfHash != confHash {
	// 	r.Log.Info("Pod config not matching with expected config. Needs rolling restart", "expectedConfHash", confHash, "currentConfHash", currentConfHash)
	// 	return true, nil
	// }

	cmName := getNamespacedNameForConfigMap(aeroCluster, rackState.Rack.ID)
	confMap := &corev1.ConfigMap{}
	err := r.Client.Get(context.TODO(), cmName, confMap)
	if err != nil {
		return false, err
	}
	requiredConfHash := confMap.Data[configmap.AerospikeConfHashFileName]
	requiredNetworkPolicyHash := confMap.Data[configmap.NetworkPolicyHashKeyFileName]

	podStatus := aeroCluster.Status.Pods[pod.Name]

	// Check if aerospikeConfig is updated
	if podStatus.AerospikeConfigHash != requiredConfHash {
		needRollingRestartPod = true
		r.Log.Info("AerospikeConfig changed. Need rolling restart",
			"requiredHash", requiredConfHash,
			"currentHash", podStatus.AerospikeConfigHash)
	}

	// Check if networkPolicy is updated
	if podStatus.NetworkPolicyHash != requiredNetworkPolicyHash {
		needRollingRestartPod = true
		r.Log.Info("Aerospike network policy changed. Need rolling restart",
			"requiredHash", requiredNetworkPolicyHash,
			"currentHash", podStatus.NetworkPolicyHash)
	}

	// Check if secret is updated
	if isAerospikeConfigSecretUpdatedInAeroCluster(aeroCluster, pod) {
		needRollingRestartPod = true
		r.Log.Info("AerospikeConfigSecret changed. Need rolling restart")
	}

	// Check if resources are updated
	if isResourceUpdatedInAeroCluster(aeroCluster, pod) {
		needRollingRestartPod = true
		r.Log.Info("Aerospike resources changed. Need rolling restart")
	}

	// Check if podSpec is updated
	if isPodSpecUpdatedInAeroCluster(aeroCluster, pod) {
		needRollingRestartPod = true
		r.Log.Info("Aerospike podSpec changed. Need rolling restart")
	}

	// Check if RACKSTORAGE/CONFIGMAP is updated
	if isRackConfigMapsUpdatedInAeroCluster(aeroCluster, rackState, pod) {
		needRollingRestartPod = true
		r.Log.Info("Aerospike rack storage configMaps changed. Need rolling restart")
	}

	return needRollingRestartPod, nil
}

func isRackConfigMapsUpdatedInAeroCluster(aeroCluster *aerospikev1alpha1.AerospikeCluster, rackState RackState, pod corev1.Pod) bool {
	configMaps, _ := rackState.Rack.Storage.GetConfigMaps()

	var volumesToBeMatched []corev1.Volume
	for _, volume := range pod.Spec.Volumes {
		if volume.ConfigMap == nil ||
			volume.Name == confDirName ||
			volume.Name == initConfDirName {
			continue
		}
		volumesToBeMatched = append(volumesToBeMatched, volume)
	}

	if len(configMaps) != len(volumesToBeMatched) {
		return true
	}

	for _, configMapVolume := range configMaps {
		// Ignore error since its not possible.
		pvcName, _ := getPVCName(configMapVolume.Path)
		found := false
		for _, volume := range volumesToBeMatched {
			if volume.Name == pvcName {
				found = true
				break
			}
		}

		if !found {
			return true
		}

	}
	return false
}

func isPodSpecUpdatedInAeroCluster(aeroCluster *aerospikev1alpha1.AerospikeCluster, pod corev1.Pod) bool {
	var extraPodContainers []corev1.Container
	for _, podContainer := range pod.Spec.Containers {
		if podContainer.Name == aerospikev1alpha1.AerospikeServerContainerName ||
			podContainer.Name == aerospikev1alpha1.AerospikeServerInitContainerName {
			// Check any other container also if added in default container list of statefulset
			continue
		}
		extraPodContainers = append(extraPodContainers, podContainer)
	}

	if len(extraPodContainers) != len(aeroCluster.Spec.PodSpec.Sidecars) {
		return true
	}
	for _, sideCar := range aeroCluster.Spec.PodSpec.Sidecars {
		found := false
		for _, podContainer := range extraPodContainers {
			if sideCar.Name == podContainer.Name {
				found = true
				break
			}
		}
		if !found {
			// SideCar not present in podSpec
			return true
		}
	}
	return false
}

func isResourceUpdatedInAeroCluster(aeroCluster *aerospikev1alpha1.AerospikeCluster, pod corev1.Pod) bool {
	res := aeroCluster.Spec.Resources
	if res == nil {
		res = &corev1.ResourceRequirements{}
	}

	if !isResourceListEqual(pod.Spec.Containers[0].Resources.Requests, res.Requests) ||
		!isResourceListEqual(pod.Spec.Containers[0].Resources.Limits, res.Limits) {
		return true
	}
	return false
}

func isAerospikeConfigSecretUpdatedInAeroCluster(aeroCluster *aerospikev1alpha1.AerospikeCluster, pod corev1.Pod) bool {
	// TODO: Should we restart when secret is removed completely
	// Can secret be even empty for enterprise cluster (feature-key is alsways required)?
	if aeroCluster.Spec.AerospikeConfigSecret.SecretName != "" {
		const secretVolumeName = "secretinfo"
		var volFound bool
		for _, vol := range pod.Spec.Volumes {
			if vol.Name == secretVolumeName {

				if vol.VolumeSource.Secret.SecretName != aeroCluster.Spec.AerospikeConfigSecret.SecretName {
					// Pod volume secret doesn't match. It needs restart
					return true
				}
				volFound = true
				break
			}
		}
		if !volFound {
			// Pod volume secret not found. It needs restart
			return true
		}

		var volmFound bool
		for _, volMount := range pod.Spec.Containers[0].VolumeMounts {
			if volMount.Name == secretVolumeName {
				if volMount.MountPath != aeroCluster.Spec.AerospikeConfigSecret.MountPath {
					// Pod volume mount doesn't match. It needs restart
					return true
				}
				volmFound = true
				break
			}
		}
		if !volmFound {
			// Pod volume mount not found. It needs restart
			return true
		}
	}
	return false
}

func (r *AerospikeClusterReconciler) rollingRestartPod(aeroCluster *aerospikev1alpha1.AerospikeCluster, rackState RackState, pod corev1.Pod) reconcileResult {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))

	// Also check if statefulSet is in stable condition
	// Check for all containers. Status.ContainerStatuses doesn't include init container
	if pod.Status.ContainerStatuses == nil {
		return reconcileError(fmt.Errorf("Pod %s containerStatus is nil, pod may be in unscheduled state", pod.Name))
	}

	r.Log.Info("Rolling restart pod", "podName", pod.Name)
	var pFound *corev1.Pod

	for i := 0; i < 5; i++ {
		r.Log.V(1).Info("Waiting for pod to be ready", "podName", pod.Name)

		pFound = &corev1.Pod{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pFound)
		if err != nil {
			r.Log.Error(err, "Failed to get pod, try retry after 5 sec")
			time.Sleep(time.Second * 5)
			continue
		}
		ps := pFound.Status.ContainerStatuses[0]
		if ps.Ready {
			break
		}

		if utils.IsCrashed(pFound) {
			r.Log.Info("Warn: Pod has crashed", "podName", pFound.Name)
			break
		}

		r.Log.Info("Warn: Pod containerStatus is not ready, try after 5 sec")
		time.Sleep(time.Second * 5)
	}

	// TODO: What if pod is still not found

	err := utils.CheckPodFailed(pFound)
	if err == nil {
		// Check for migration
		if res := r.waitForNodeSafeStopReady(aeroCluster, pFound); !res.isSuccess {
			return res
		}
	} else {
		// TODO: Check a user flag to restart failed pods.
		r.Log.Info("Restarting failed pod", "podName", pFound.Name, "error", err)
	}

	// Delete pod
	if err := r.Client.Delete(context.TODO(), pFound); err != nil {
		r.Log.Error(err, "Failed to delete pod")
		return reconcileError(err)
	}
	r.Log.V(1).Info("Pod deleted", "podName", pFound.Name)

	// Wait for pod to come up
	var started bool
	for i := 0; i < 20; i++ {
		r.Log.V(1).Info("Waiting for pod to be ready after delete", "podName", pFound.Name, "status", pFound.Status.Phase, "DeletionTimestamp", pFound.DeletionTimestamp)

		pod := &corev1.Pod{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: pFound.Name, Namespace: pFound.Namespace}, pod)
		if err != nil {
			r.Log.Error(err, "Failed to get pod")
			time.Sleep(time.Second * 5)
			continue
		}

		if err := utils.CheckPodFailed(pod); err != nil {
			return reconcileError(err)
		}

		if !utils.IsPodRunningAndReady(pod) {
			r.Log.V(1).Info("Waiting for pod to be ready", "podName", pod.Name, "status", pod.Status.Phase, "DeletionTimestamp", pod.DeletionTimestamp)
			time.Sleep(time.Second * 5)
			continue
		}

		r.Log.Info("Pod is restarted", "podName", pod.Name)
		started = true
		break
	}

	// TODO: In what situation this can happen?
	if !started {
		r.Log.Info("Warn: Pos is not running or ready. Pod might also be terminating", "podName", pod.Name, "status", pod.Status.Phase, "DeletionTimestamp", pod.DeletionTimestamp)
	}

	return reconcileSuccess()
}

func (r *AerospikeClusterReconciler) scaleDownRack(aeroCluster *aerospikev1alpha1.AerospikeCluster, found *appsv1.StatefulSet, rackState RackState) (*appsv1.StatefulSet, reconcileResult) {
	// logger := pkglog.New("AerospikeClusterSTS", getNamespacedNameForStatefulSet(aeroCluster, rackState.Rack.ID))

	desiredSize := int32(rackState.Size)

	r.Log.Info("ScaleDown AerospikeCluster statefulset", "desiredSz", desiredSize, "currentSz", *found.Spec.Replicas)

	// Continue if scaleDown is not needed
	if *found.Spec.Replicas <= desiredSize {
		return found, reconcileSuccess()
	}

	oldPodList, err := r.getRackPodList(aeroCluster, rackState.Rack.ID)
	if err != nil {
		return found, reconcileError(fmt.Errorf("Failed to list pods: %v", err))
	}

	if r.isAnyPodInFailedState(aeroCluster, oldPodList.Items) {
		return found, reconcileError(fmt.Errorf("Cannot scale down AerospikeCluster. A pod is already in failed state"))
	}

	var pod *corev1.Pod

	if *found.Spec.Replicas > desiredSize {

		// maintain list of removed pods. It will be used for alumni-reset and tip-clear
		podName := getStatefulSetPodName(found.Name, *found.Spec.Replicas-1)

		pod = utils.GetPod(podName, oldPodList.Items)

		// Ignore safe stop check on pod not in running state.
		if utils.IsPodRunningAndReady(pod) {
			if res := r.waitForNodeSafeStopReady(aeroCluster, pod); !res.isSuccess {
				// The pod is running and is unsafe to terminate.
				return found, res
			}
		}

		// Update new object with new size
		newSize := *found.Spec.Replicas - 1
		found.Spec.Replicas = &newSize
		if err := r.Client.Update(context.TODO(), found, updateOption); err != nil {
			return found, reconcileError(fmt.Errorf("Failed to update pod size %d StatefulSet pods: %v", newSize, err))
		}

		// Wait for pods to get terminated
		if err := r.waitForStatefulSetToBeReady(found); err != nil {
			return found, reconcileError(fmt.Errorf("Failed to wait for statefulset to be ready: %v", err))
		}

		// Fetch new object
		nFound, err := r.getStatefulSet(aeroCluster, rackState)
		if err != nil {
			return found, reconcileError(fmt.Errorf("Failed to get StatefulSet pods: %v", err))
		}
		found = nFound

		err = r.cleanupPods(aeroCluster, []string{podName}, rackState)
		if err != nil {
			return nFound, reconcileError(fmt.Errorf("Failed to cleanup pod %s: %v", podName, err))
		}

		r.Log.Info("Pod Removed", "podName", podName)
	}

	newPodList, err := r.getRackPodList(aeroCluster, rackState.Rack.ID)
	if err != nil {
		return found, reconcileError(fmt.Errorf("Failed to list pods: %v", err))
	}

	// Do post-remove-node info calls
	for _, np := range newPodList.Items {
		// TODO: We remove node from the end. Nodes will not have seed of successive nodes
		// So this will be no op.
		// We should tip in all nodes the same seed list,
		// then only this will have any impact. Is it really necessary?

		// TODO: tip after scaleup and create
		// All nodes from other rack
		r.tipClearHostname(aeroCluster, &np, pod)

		r.alumniReset(aeroCluster, &np)
	}

	if aeroCluster.Spec.MultiPodPerHost {
		// Remove service for pod
		// TODO: make it more roboust, what if it fails
		if err := r.deleteServiceForPod(pod.Name, aeroCluster.Namespace); err != nil {
			return found, reconcileError(err)
		}
	}

	return found, reconcileRequeueAfter(0)
}

func (r *AerospikeClusterReconciler) reconcileAccessControl(aeroCluster *aerospikev1alpha1.AerospikeCluster) error {
	logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))

	config, err := aerospikev1alpha1.ToAeroConfMap(aeroCluster.Spec.AerospikeConfig)
	if err != nil {
		return err
	}
	enabled, err := aerospikev1alpha1.IsSecurityEnabled(config)
	if err != nil {
		return fmt.Errorf("Failed to get cluster security status: %v", err)
	}
	if !enabled {
		r.Log.Info("Cluster is not security enabled, please enable security for this cluster.")
		return nil
	}

	// Create client
	conns, err := r.newAllHostConn(aeroCluster)
	if err != nil {
		return fmt.Errorf("Failed to get host info: %v", err)
	}
	var hosts []*as.Host
	for _, conn := range conns {
		hosts = append(hosts, &as.Host{
			Name:    conn.ASConn.AerospikeHostName,
			TLSName: conn.ASConn.AerospikeTLSName,
			Port:    conn.ASConn.AerospikePort,
		})
	}
	// Create policy using status, status has current connection info
	clientPolicy := r.getClientPolicy(aeroCluster)
	aeroClient, err := as.NewClientWithPolicyAndHost(clientPolicy, hosts...)

	if err != nil {
		return fmt.Errorf("Failed to create aerospike cluster client: %v", err)
	}

	defer aeroClient.Close()

	pp := r.getPasswordProvider(aeroCluster)
	err = accessControl.ReconcileAccessControl(&aeroCluster.Spec, &aeroCluster.Status.AerospikeClusterSpec, aeroClient, pp, logger)
	return err
}

func (r *AerospikeClusterReconciler) updateStatus(aeroCluster *aerospikev1alpha1.AerospikeCluster) error {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))

	r.Log.Info("Update status for AerospikeCluster")

	// Get the old object, it may have been updated in between.
	newAeroCluster := &aerospikev1alpha1.AerospikeCluster{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: aeroCluster.Name, Namespace: aeroCluster.Namespace}, newAeroCluster)
	if err != nil {
		return err
	}

	// Deep copy merges so blank out the spec part of status before copying over.
	newAeroCluster.Status.AerospikeClusterSpec = aerospikev1alpha1.AerospikeClusterSpec{}
	if err := lib.DeepCopy(&newAeroCluster.Status.AerospikeClusterSpec, &aeroCluster.Spec); err != nil {
		return err
	}

	err = r.patchStatus(aeroCluster, newAeroCluster)
	if err != nil {
		return fmt.Errorf("Error updating status: %v", err)
	}
	r.Log.Info("Updated status", "status", newAeroCluster.Status)
	return nil
}

func (r *AerospikeClusterReconciler) createStatus(aeroCluster *aerospikev1alpha1.AerospikeCluster) error {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))

	r.Log.Info("Creating status for AerospikeCluster")

	// Get the old object, it may have been updated in between.
	newAeroCluster := &aerospikev1alpha1.AerospikeCluster{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: aeroCluster.Name, Namespace: aeroCluster.Namespace}, newAeroCluster)
	if err != nil {
		return err
	}

	if newAeroCluster.Status.Pods == nil {
		newAeroCluster.Status.Pods = map[string]aerospikev1alpha1.AerospikePodStatus{}
	}

	if err = r.Client.Status().Update(context.TODO(), newAeroCluster); err != nil {
		return fmt.Errorf("Error creating status: %v", err)
	}

	return nil
}

func (r *AerospikeClusterReconciler) isNewCluster(aeroCluster *aerospikev1alpha1.AerospikeCluster) (bool, error) {
	if aeroCluster.Status.AerospikeConfig.Raw != nil {
		// We have valid status, cluster cannot be new.
		return false, nil
	}

	statefulSetList, err := r.getClusterStatefulSets(aeroCluster)

	if err != nil {
		return false, err
	}

	// Cluster can have status nil and still have pods on failures.
	// For cluster to be new there should be no pods in the cluster.
	return len(statefulSetList.Items) == 0, nil
}

func (r *AerospikeClusterReconciler) hasClusterFailed(aeroCluster *aerospikev1alpha1.AerospikeCluster) (bool, error) {
	isNew, err := r.isNewCluster(aeroCluster)

	if err != nil {
		// Checking cluster status failed.
		return false, err
	}

	return !isNew && aeroCluster.Status.AerospikeConfig.Raw == nil, nil
}

func (r *AerospikeClusterReconciler) patchStatus(oldAeroCluster, newAeroCluster *aerospikev1alpha1.AerospikeCluster) error {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(oldAeroCluster))

	oldJSON, err := json.Marshal(oldAeroCluster)
	if err != nil {
		return fmt.Errorf("Error marshalling old status: %v", err)
	}

	newJSON, err := json.Marshal(newAeroCluster)
	if err != nil {
		return fmt.Errorf("Error marshalling new status: %v", err)
	}

	jsonpatchPatch, err := jsonpatch.CreatePatch(oldJSON, newJSON)
	if err != nil {
		return fmt.Errorf("Error creating json patch: %v", err)
	}

	// Pick changes to the status object only.
	filteredPatch := []jsonpatch.JsonPatchOperation{}
	for _, operation := range jsonpatchPatch {
		// pods should never be updated here
		// pods is updated only from 2 places
		// 1: While pod init, it will add pod in pods
		// 2: While pod cleanup, it will remove pod from pods
		if strings.HasPrefix(operation.Path, "/status") && !strings.HasPrefix(operation.Path, "/status/pods") {
			filteredPatch = append(filteredPatch, operation)
		}
	}

	if len(filteredPatch) == 0 {
		r.Log.Info("No status change required")
		return nil
	}
	r.Log.V(1).Info("Filtered status patch ", "patch", filteredPatch, "oldObj.status", oldAeroCluster.Status, "newObj.status", newAeroCluster.Status)

	jsonpatchJSON, err := json.Marshal(filteredPatch)

	if err != nil {
		return fmt.Errorf("Error marshalling json patch: %v", err)
	}

	patch := client.RawPatch(types.JSONPatchType, jsonpatchJSON)

	if err = r.Client.Status().Patch(context.TODO(), oldAeroCluster, patch, client.FieldOwner(patchFieldOwner)); err != nil {
		return fmt.Errorf("Error patching status: %v", err)
	}

	return nil
	// FIXME: Json unmarshal used by above client.Status(),Patch()  does not convert empty lists in the new JSON to empty lists in the target. Seems like a bug in encoding/json/Unmarshall.
	//
	// Workaround by force copying new object's status to old object's status.
	// return lib.DeepCopy(&oldAeroCluster.Status, &newAeroCluster.Status)
}

// removePodStatus removes podNames from the cluster's pod status.
// Assumes the pods are not running so that the no concurrent update to this pod status is possbile.
func (r *AerospikeClusterReconciler) removePodStatus(aeroCluster *aerospikev1alpha1.AerospikeCluster, podNames []string) error {
	if len(podNames) == 0 {
		return nil
	}

	patches := []jsonpatch.JsonPatchOperation{}

	for _, podName := range podNames {
		patch := jsonpatch.JsonPatchOperation{
			Operation: "remove",
			Path:      "/status/pods/" + podName,
		}
		patches = append(patches, patch)
	}

	jsonpatchJSON, err := json.Marshal(patches)
	constantPatch := client.RawPatch(types.JSONPatchType, jsonpatchJSON)

	// Since the pod status is updated from pod init container, set the fieldowner to "pod" for pod status updates.
	if err = r.Client.Status().Patch(context.TODO(), aeroCluster, constantPatch, client.FieldOwner("pod")); err != nil {
		return fmt.Errorf("Error updating status: %v", err)
	}

	return nil
}

// recoverFailedCreate deletes the stateful sets for every rack and retries creating the cluster again when the first cluster create has failed.
//
// The cluster is not new but maybe unreachable or down. There could be an Aerospike configuration
// error that passed the operator validation but is invalid on the server. This will happen for
// example where deeper paramter or value of combination of parameter values need validation which
// is missed by the operator. For e.g. node-address-port values in xdr datacenter section needs better
// validation for ip and port.
//
// Such cases warrant a cluster recreate to recover after the user corrects the configuration.
func (r *AerospikeClusterReconciler) recoverFailedCreate(aeroCluster *aerospikev1alpha1.AerospikeCluster) error {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))
	r.Log.Info("Forcing a cluster recreate as status is nil. The cluster could be unreachable due to bad configuration.")

	// Delete all statefulsets and everything related so that it can be properly created and updated in next run.
	statefulSetList, err := r.getClusterStatefulSets(aeroCluster)
	if err != nil {
		return fmt.Errorf("Error getting statefulsets while forcing recreate of the cluster as status is nil: %v", err)
	}

	r.Log.V(1).Info("Found statefulset for cluster. Need to delete them", "nSTS", len(statefulSetList.Items))
	for _, statefulset := range statefulSetList.Items {
		if err := r.deleteStatefulSet(aeroCluster, &statefulset); err != nil {
			return fmt.Errorf("Error deleting statefulset while forcing recreate of the cluster as status is nil: %v", err)
		}
	}

	// Clear pod status as well in status since we want to be re-initializing or cascade deleting devices if any.
	// This is not necessary since scale-up would cleanup danglin pod status. However done here for general
	// cleanliness.
	rackStateList := getNewRackStateList(aeroCluster)
	for _, state := range rackStateList {
		pods, err := r.getRackPodList(aeroCluster, state.Rack.ID)
		if err != nil {
			return fmt.Errorf("Failed recover failed cluster: %v", err)
		}

		newPodNames := []string{}
		for i := 0; i < len(pods.Items); i++ {
			newPodNames = append(newPodNames, pods.Items[i].Name)
		}

		err = r.cleanupPods(aeroCluster, newPodNames, state)
		if err != nil {
			return fmt.Errorf("Failed recover failed cluster: %v", err)
		}
	}

	return fmt.Errorf("Forcing recreate of the cluster as status is nil")
}

// cleanupPods checks pods and status before scaleup to detect and fix any status anomalies.
func (r *AerospikeClusterReconciler) cleanupPods(aeroCluster *aerospikev1alpha1.AerospikeCluster, podNames []string, rackState RackState) error {
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))

	needStatusCleanup := []string{}
	for _, podName := range podNames {
		_, ok := aeroCluster.Status.Pods[podName]
		if ok {
			needStatusCleanup = append(needStatusCleanup, podName)
		}
	}

	if len(needStatusCleanup) > 0 {
		r.Log.Info("Removing pod status for dangling pods", "pods", podNames)

		if err := r.removePodStatus(aeroCluster, needStatusCleanup); err != nil {
			return fmt.Errorf("Could not cleanup pod status: %v", err)
		}
	}

	r.Log.Info("Removing pvc for removed pods", "pods", podNames)

	// Delete PVCs if cascadeDelete
	pvcItems, err := r.getPodsPVCList(aeroCluster, podNames, rackState.Rack.ID)
	if err != nil {
		return fmt.Errorf("Could not find pvc for pods %v: %v", podNames, err)
	}
	storage := rackState.Rack.Storage
	if err := r.removePVCs(aeroCluster, &storage, pvcItems); err != nil {
		return fmt.Errorf("Could not cleanup pod PVCs: %v", err)
	}

	return nil
}

func (r *AerospikeClusterReconciler) addFinalizer(aeroCluster *aerospikev1alpha1.AerospikeCluster, finalizerName string) error {
	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent
	// registering our finalizer.

	if !containsString(aeroCluster.ObjectMeta.Finalizers, finalizerName) {
		r.Log.V(1).Info("Adding finalizer", "Current finalizers", aeroCluster.ObjectMeta.Finalizers, "Added", finalizerName)

		aeroCluster.ObjectMeta.Finalizers = append(aeroCluster.ObjectMeta.Finalizers, finalizerName)
		if err := r.Client.Update(context.TODO(), aeroCluster); err != nil {
			return err
		}
	}
	return nil
}

func (r *AerospikeClusterReconciler) cleanUpAndremoveFinalizer(aeroCluster *aerospikev1alpha1.AerospikeCluster, finalizerName string) error {
	// The object is being deleted
	if containsString(aeroCluster.ObjectMeta.Finalizers, finalizerName) {
		// Handle any external dependency
		if err := r.deleteExternalResources(aeroCluster); err != nil {
			// If fail to delete the external dependency here, return with error
			// so that it can be retried
			return err
		}

		// Remove finalizer from the list
		aeroCluster.ObjectMeta.Finalizers = removeString(aeroCluster.ObjectMeta.Finalizers, finalizerName)
		if err := r.Client.Update(context.TODO(), aeroCluster); err != nil {
			return err
		}
	}

	// Stop reconciliation as the item is being deleted
	return nil
}

func (r *AerospikeClusterReconciler) deleteExternalResources(aeroCluster *aerospikev1alpha1.AerospikeCluster) error {
	// Delete should be idempotent
	// logger := pkglog.New("AerospikeCluster", utils.ClusterNamespacedName(aeroCluster))

	r.Log.Info("Removing pvc for removed cluster")

	// Delete pvc for all rack storage
	for _, rack := range aeroCluster.Spec.RackConfig.Racks {
		rackPVCItems, err := r.getRackPVCList(aeroCluster, rack.ID)
		if err != nil {
			return fmt.Errorf("Could not find pvc for rack: %v", err)
		}
		storage := rack.Storage
		if _, err := r.removePVCsAsync(aeroCluster, &storage, rackPVCItems); err != nil {
			return fmt.Errorf("Failed to remove cluster PVCs: %v", err)
		}
	}

	// Delete PVCs for any remaining old removed racks
	pvcItems, err := r.getClusterPVCList(aeroCluster)
	if err != nil {
		return fmt.Errorf("Could not find pvc for cluster: %v", err)
	}

	// removePVCs should be passed only filtered pvc otherwise rack pvc may be removed using global storage cascadeDelete
	var fileredPVCItems []corev1.PersistentVolumeClaim
	for _, pvc := range pvcItems {
		var found bool
		for _, rack := range aeroCluster.Spec.RackConfig.Racks {
			rackLables := utils.LabelsForAerospikeClusterRack(aeroCluster.Name, rack.ID)
			if reflect.DeepEqual(pvc.Labels, rackLables) {
				found = true
				break
			}
		}
		if !found {
			fileredPVCItems = append(fileredPVCItems, pvc)
		}
	}

	// Delete pvc for commmon storage.
	if _, err := r.removePVCsAsync(aeroCluster, &aeroCluster.Spec.Storage, fileredPVCItems); err != nil {
		return fmt.Errorf("Failed to remove cluster PVCs: %v", err)
	}

	return nil
}

func (r *AerospikeClusterReconciler) removePVCs(aeroCluster *aerospikev1alpha1.AerospikeCluster, storage *aerospikev1alpha1.AerospikeStorageSpec, pvcItems []corev1.PersistentVolumeClaim) error {
	deletedPVCs, err := r.removePVCsAsync(aeroCluster, storage, pvcItems)
	if err != nil {
		return err
	}

	return r.waitForPVCTermination(aeroCluster, deletedPVCs)
}

func (r *AerospikeClusterReconciler) removePVCsAsync(aeroCluster *aerospikev1alpha1.AerospikeCluster, storage *aerospikev1alpha1.AerospikeStorageSpec, pvcItems []corev1.PersistentVolumeClaim) ([]corev1.PersistentVolumeClaim, error) {
	aeroClusterNamespacedName := getNamespacedNameForCluster(aeroCluster)

	// logger := pkglog.New("AerospikeCluster", aeroClusterNamespacedName)

	deletedPVCs := []corev1.PersistentVolumeClaim{}

	for _, pvc := range pvcItems {
		if utils.IsPVCTerminating(&pvc) {
			continue
		}
		// Should we wait for delete?
		// Can we do it async in scaleDown

		// Check for path in pvc annotations. We put path annotation while creating statefulset
		path, ok := pvc.Annotations[storagePathAnnotationKey]
		if !ok {
			r.Log.Info("Warn: PVC can not be removed, it does not have storage-path annotation", "PVC", pvc.Name, "annotations", pvc.Annotations)
			continue
		}

		var cascadeDelete bool
		v := getVolumeConfigForPVC(aeroClusterNamespacedName.Name, storage, path)
		if v == nil {
			if *pvc.Spec.VolumeMode == corev1.PersistentVolumeBlock {
				cascadeDelete = storage.BlockVolumePolicy.CascadeDelete
			} else {
				cascadeDelete = storage.FileSystemVolumePolicy.CascadeDelete
			}
			r.Log.Info("PVC path not found in configured storage volumes. Use storage level cascadeDelete policy", "PVC", pvc.Name, "path", path, "cascadeDelete", cascadeDelete)

		} else {
			cascadeDelete = v.CascadeDelete
		}

		if cascadeDelete {
			deletedPVCs = append(deletedPVCs, pvc)
			if err := r.Client.Delete(context.TODO(), &pvc); err != nil {
				return nil, fmt.Errorf("Could not delete pvc %s: %v", pvc.Name, err)
			}
			r.Log.Info("PVC removed", "PVC", pvc.Name, "PVCCascadeDelete", cascadeDelete)
		} else {
			r.Log.Info("PVC not removed", "PVC", pvc.Name, "PVCCascadeDelete", cascadeDelete)
		}
	}

	return deletedPVCs, nil
}

func (r *AerospikeClusterReconciler) waitForPVCTermination(aeroCluster *aerospikev1alpha1.AerospikeCluster, deletedPVCs []corev1.PersistentVolumeClaim) error {
	if len(deletedPVCs) == 0 {
		return nil
	}

	// aeroClusterNamespacedName := getNamespacedNameForCluster(aeroCluster)

	// logger := pkglog.New("AerospikeCluster", aeroClusterNamespacedName)

	// Wait for the PVCs to actually be deleted.
	pollAttempts := 15
	sleepInterval := time.Second * 20

	pending := false
	for i := 0; i < pollAttempts; i++ {
		pending = false
		existingPVCs, err := r.getClusterPVCList(aeroCluster)
		if err != nil {
			return err
		}

		for _, pvc := range deletedPVCs {
			found := false
			for _, existing := range existingPVCs {
				if existing.Name == pvc.Name {
					r.Log.Info("Waiting for PVC termination", "PVC", pvc.Name)
					found = true
					break
				}
			}

			if found {
				pending = true
				break
			}
		}

		if !pending {
			// All to-delete PVCs are deleted.
			break
		}

		// Wait for some more time.
		time.Sleep(sleepInterval)
	}

	if pending {
		return fmt.Errorf("PVC termination timed out PVC: %v", deletedPVCs)
	}

	return nil
}

func getVolumeConfigForPVC(aeroClusterName string, storage *aerospikev1alpha1.AerospikeStorageSpec, pvcPathAnnotation string) *aerospikev1alpha1.AerospikePersistentVolumeSpec {
	volumes := storage.Volumes
	for _, v := range volumes {
		if pvcPathAnnotation == v.Path {
			return &v
		}
	}
	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
