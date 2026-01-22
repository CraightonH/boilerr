/*
Copyright 2026 CraightonH.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
	"github.com/CraightonH/boilerr/internal/config"
	"github.com/CraightonH/boilerr/internal/resources"
)

const (
	// FinalizerName is the finalizer used by the SteamServer controller.
	FinalizerName = "boilerr.dev/steamserver-finalizer"
)

// SteamServerReconciler reconciles a SteamServer object.
type SteamServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=boilerr.dev,resources=steamservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=boilerr.dev,resources=steamservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=boilerr.dev,resources=steamservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=boilerr.dev,resources=gamedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is the main reconciliation loop for SteamServer resources.
func (r *SteamServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the SteamServer CR
	server := &boilerrv1alpha1.SteamServer{}
	if err := r.Get(ctx, req.NamespacedName, server); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("SteamServer resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get SteamServer")
		return ctrl.Result{}, err
	}

	// 2. Handle deletion with finalizer
	if !server.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, server)
	}

	// 3. Add finalizer if not present
	if !controllerutil.ContainsFinalizer(server, FinalizerName) {
		controllerutil.AddFinalizer(server, FinalizerName)
		if err := r.Update(ctx, server); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 4. Fetch the referenced GameDefinition (cluster-scoped, no namespace)
	gameDef, err := r.fetchGameDefinition(ctx, server)
	if err != nil {
		return r.setErrorStatus(ctx, server, "GameDefinition", err)
	}

	// 5. Check GameDefinition is ready
	if gameDef != nil && !gameDef.Status.Ready {
		err := fmt.Errorf("GameDefinition %q is not ready: %s", server.Spec.GameDefinition, gameDef.Status.Message)
		if _, statusErr := r.setErrorStatus(ctx, server, "GameDefinition", err); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// 6. Validate config against schema
	if gameDef != nil && gameDef.Spec.ConfigSchema != nil {
		if err := config.ValidateConfig(server.Spec.Config, gameDef.Spec.ConfigSchema); err != nil {
			return r.setErrorStatus(ctx, server, "Config", err)
		}
	}

	// 7. Reconcile child resources
	if err := r.reconcileConfigMap(ctx, server, gameDef); err != nil {
		return r.setErrorStatus(ctx, server, "ConfigMap", err)
	}

	if err := r.reconcilePVC(ctx, server, gameDef); err != nil {
		return r.setErrorStatus(ctx, server, "PVC", err)
	}

	if err := r.reconcileStatefulSet(ctx, server, gameDef); err != nil {
		return r.setErrorStatus(ctx, server, "StatefulSet", err)
	}

	if err := r.reconcileService(ctx, server, gameDef); err != nil {
		return r.setErrorStatus(ctx, server, "Service", err)
	}

	// 8. Update status based on actual state
	return r.updateStatus(ctx, server)
}

// fetchGameDefinition fetches the GameDefinition referenced by the SteamServer.
// Returns nil if no game is specified (fallback mode).
func (r *SteamServerReconciler) fetchGameDefinition(ctx context.Context, server *boilerrv1alpha1.SteamServer) (*boilerrv1alpha1.GameDefinition, error) {
	if server.Spec.GameDefinition == "" {
		// Fallback mode: no GameDefinition, all values from SteamServer spec
		return nil, nil
	}

	var gameDef boilerrv1alpha1.GameDefinition
	// GameDefinition is cluster-scoped, so no namespace in the key
	if err := r.Get(ctx, client.ObjectKey{Name: server.Spec.GameDefinition}, &gameDef); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("GameDefinition %q not found", server.Spec.GameDefinition)
		}
		return nil, err
	}

	return &gameDef, nil
}

// handleDeletion handles the deletion of a SteamServer resource.
func (r *SteamServerReconciler) handleDeletion(ctx context.Context, server *boilerrv1alpha1.SteamServer) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(server, FinalizerName) {
		logger.Info("Running finalizer cleanup for SteamServer")

		// Perform cleanup if needed
		// Child resources with owner references will be garbage collected automatically

		// Remove finalizer
		controllerutil.RemoveFinalizer(server, FinalizerName)
		if err := r.Update(ctx, server); err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// reconcileConfigMap ensures the ConfigMap exists if config files are specified.
func (r *SteamServerReconciler) reconcileConfigMap(ctx context.Context, server *boilerrv1alpha1.SteamServer, _ *boilerrv1alpha1.GameDefinition) error {
	logger := log.FromContext(ctx)

	// Skip if no config files
	if len(server.Spec.ConfigFiles) == 0 {
		return nil
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.ConfigMapName(server.Name),
			Namespace: server.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Labels = commonLabels(server.Name, server.Spec.GameDefinition)
		configMap.Data = make(map[string]string)

		for i, cf := range server.Spec.ConfigFiles {
			key := fmt.Sprintf("config-%d", i)
			configMap.Data[key] = cf.Content
		}

		return controllerutil.SetControllerReference(server, configMap, r.Scheme)
	})

	if err != nil {
		return err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("ConfigMap reconciled", "operation", op)
	}

	return nil
}

// reconcilePVC ensures the PVC exists for the SteamServer.
func (r *SteamServerReconciler) reconcilePVC(ctx context.Context, server *boilerrv1alpha1.SteamServer, gameDef *boilerrv1alpha1.GameDefinition) error {
	logger := log.FromContext(ctx)

	pvcBuilder := resources.NewPVCBuilder(server, gameDef)
	desiredPVC := pvcBuilder.Build()

	if desiredPVC == nil {
		// No storage configured
		return nil
	}

	existingPVC := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      desiredPVC.Name,
		Namespace: desiredPVC.Namespace,
	}, existingPVC)

	if apierrors.IsNotFound(err) {
		// Create new PVC
		if err := controllerutil.SetControllerReference(server, desiredPVC, r.Scheme); err != nil {
			return err
		}

		logger.Info("Creating PVC", "name", desiredPVC.Name)
		return r.Create(ctx, desiredPVC)
	} else if err != nil {
		return err
	}

	// PVC exists - PVCs are immutable for most fields, so we only update labels
	if !hasLabels(existingPVC.Labels, desiredPVC.Labels) {
		existingPVC.Labels = desiredPVC.Labels
		logger.Info("Updating PVC labels", "name", existingPVC.Name)
		return r.Update(ctx, existingPVC)
	}

	return nil
}

// reconcileStatefulSet ensures the StatefulSet exists and is up to date.
func (r *SteamServerReconciler) reconcileStatefulSet(ctx context.Context, server *boilerrv1alpha1.SteamServer, gameDef *boilerrv1alpha1.GameDefinition) error {
	logger := log.FromContext(ctx)

	stsBuilder := resources.NewStatefulSetBuilder(server, gameDef)
	desiredSTS := stsBuilder.Build()

	// Set owner reference before create/update
	if err := controllerutil.SetControllerReference(server, desiredSTS, r.Scheme); err != nil {
		return err
	}

	existingSTS := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      desiredSTS.Name,
		Namespace: desiredSTS.Namespace,
	}, existingSTS)

	if apierrors.IsNotFound(err) {
		logger.Info("Creating StatefulSet", "name", desiredSTS.Name)
		return r.Create(ctx, desiredSTS)
	} else if err != nil {
		return err
	}

	// Update the StatefulSet spec
	existingSTS.Spec = desiredSTS.Spec
	existingSTS.Labels = desiredSTS.Labels

	logger.Info("Updating StatefulSet", "name", existingSTS.Name)
	return r.Update(ctx, existingSTS)
}

// reconcileService ensures the Service exists and is up to date.
func (r *SteamServerReconciler) reconcileService(ctx context.Context, server *boilerrv1alpha1.SteamServer, gameDef *boilerrv1alpha1.GameDefinition) error {
	logger := log.FromContext(ctx)

	svcBuilder := resources.NewServiceBuilder(server, gameDef)
	desiredSVC := svcBuilder.Build()

	existingSVC := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      desiredSVC.Name,
		Namespace: desiredSVC.Namespace,
	}, existingSVC)

	if apierrors.IsNotFound(err) {
		// Set owner reference
		if err := controllerutil.SetControllerReference(server, desiredSVC, r.Scheme); err != nil {
			return err
		}

		logger.Info("Creating Service", "name", desiredSVC.Name)
		return r.Create(ctx, desiredSVC)
	} else if err != nil {
		return err
	}

	// Update service - preserve ClusterIP if already assigned
	existingSVC.Spec.Type = desiredSVC.Spec.Type
	existingSVC.Spec.Selector = desiredSVC.Spec.Selector
	existingSVC.Spec.Ports = desiredSVC.Spec.Ports
	existingSVC.Labels = desiredSVC.Labels

	logger.Info("Updating Service", "name", existingSVC.Name)
	return r.Update(ctx, existingSVC)
}

// updateStatus updates the SteamServer status based on the actual cluster state.
func (r *SteamServerReconciler) updateStatus(ctx context.Context, server *boilerrv1alpha1.SteamServer) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the StatefulSet to determine server state
	sts := &appsv1.StatefulSet{}
	stsErr := r.Get(ctx, client.ObjectKey{
		Name:      server.Name,
		Namespace: server.Namespace,
	}, sts)

	// Get the Service to determine address
	svc := &corev1.Service{}
	svcErr := r.Get(ctx, client.ObjectKey{
		Name:      server.Name,
		Namespace: server.Namespace,
	}, svc)

	// Determine state
	newState := r.determineState(ctx, server, sts, stsErr)
	newAddress := r.determineAddress(svc, svcErr)
	newPorts := r.determinePorts(svc, svcErr)
	now := metav1.Now()

	// Check if status needs update
	statusChanged := server.Status.State != newState ||
		server.Status.Address != newAddress ||
		!portsEqual(server.Status.Ports, newPorts)

	if statusChanged {
		server.Status.State = newState
		server.Status.Address = newAddress
		server.Status.Ports = newPorts
		server.Status.LastUpdated = &now
		server.Status.Message = r.stateMessage(newState)

		logger.Info("Updating SteamServer status",
			"state", newState,
			"address", newAddress,
		)

		if err := r.Status().Update(ctx, server); err != nil {
			logger.Error(err, "Failed to update SteamServer status")
			return ctrl.Result{}, err
		}
	}

	// Requeue if not yet running to check for state changes
	if newState != boilerrv1alpha1.ServerStateRunning {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// determineState determines the current state of the server based on the StatefulSet.
func (r *SteamServerReconciler) determineState(ctx context.Context, server *boilerrv1alpha1.SteamServer, sts *appsv1.StatefulSet, stsErr error) boilerrv1alpha1.ServerState {
	if stsErr != nil {
		if apierrors.IsNotFound(stsErr) {
			return boilerrv1alpha1.ServerStatePending
		}
		return boilerrv1alpha1.ServerStateError
	}

	// Check StatefulSet replica status
	if sts.Status.ReadyReplicas == 0 && sts.Status.Replicas == 0 {
		return boilerrv1alpha1.ServerStatePending
	}

	// Check if pod exists and its state
	pod := &corev1.Pod{}
	podName := fmt.Sprintf("%s-0", server.Name)
	podErr := r.Get(ctx, client.ObjectKey{
		Name:      podName,
		Namespace: server.Namespace,
	}, pod)

	if podErr != nil {
		if apierrors.IsNotFound(podErr) {
			return boilerrv1alpha1.ServerStatePending
		}
		return boilerrv1alpha1.ServerStateError
	}

	// Check pod phase
	switch pod.Status.Phase {
	case corev1.PodPending:
		return boilerrv1alpha1.ServerStatePending
	case corev1.PodRunning:
		// Check init containers
		for _, cs := range pod.Status.InitContainerStatuses {
			if cs.Name == resources.InitContainerName {
				if cs.State.Running != nil {
					return boilerrv1alpha1.ServerStateInstalling
				}
				if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
					return boilerrv1alpha1.ServerStateError
				}
			}
		}
		// Check main container
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == resources.GameServerContainerName {
				if cs.State.Running != nil && cs.Ready {
					return boilerrv1alpha1.ServerStateRunning
				}
				if cs.State.Running != nil {
					return boilerrv1alpha1.ServerStateStarting
				}
			}
		}
		return boilerrv1alpha1.ServerStateStarting
	case corev1.PodFailed:
		return boilerrv1alpha1.ServerStateError
	default:
		return boilerrv1alpha1.ServerStatePending
	}
}

// determineAddress determines the external address from the Service.
func (r *SteamServerReconciler) determineAddress(svc *corev1.Service, svcErr error) string {
	if svcErr != nil {
		return ""
	}

	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				return ingress.IP
			}
			if ingress.Hostname != "" {
				return ingress.Hostname
			}
		}
		return "<pending>"
	case corev1.ServiceTypeNodePort:
		return "<node-ip>"
	case corev1.ServiceTypeClusterIP:
		return svc.Spec.ClusterIP
	default:
		return ""
	}
}

// determinePorts extracts port information from the Service.
func (r *SteamServerReconciler) determinePorts(svc *corev1.Service, svcErr error) []boilerrv1alpha1.PortStatus {
	if svcErr != nil {
		return nil
	}

	ports := make([]boilerrv1alpha1.PortStatus, len(svc.Spec.Ports))
	for i, p := range svc.Spec.Ports {
		port := p.Port
		if svc.Spec.Type == corev1.ServiceTypeNodePort && p.NodePort != 0 {
			port = p.NodePort
		}
		ports[i] = boilerrv1alpha1.PortStatus{
			Name:     p.Name,
			Port:     port,
			Protocol: p.Protocol,
		}
	}

	return ports
}

// setErrorStatus sets the server status to Error with a message.
func (r *SteamServerReconciler) setErrorStatus(ctx context.Context, server *boilerrv1alpha1.SteamServer, resource string, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Error(err, "Failed to reconcile resource", "resource", resource)

	now := metav1.Now()
	server.Status.State = boilerrv1alpha1.ServerStateError
	server.Status.Message = fmt.Sprintf("Failed to reconcile %s: %v", resource, err)
	server.Status.LastUpdated = &now

	if statusErr := r.Status().Update(ctx, server); statusErr != nil {
		logger.Error(statusErr, "Failed to update error status")
		return ctrl.Result{}, statusErr
	}

	return ctrl.Result{}, err
}

// stateMessage returns a human-readable message for a state.
func (r *SteamServerReconciler) stateMessage(state boilerrv1alpha1.ServerState) string {
	switch state {
	case boilerrv1alpha1.ServerStatePending:
		return "Waiting for resources to be scheduled"
	case boilerrv1alpha1.ServerStateInstalling:
		return "SteamCMD is downloading game files"
	case boilerrv1alpha1.ServerStateStarting:
		return "Game server is starting up"
	case boilerrv1alpha1.ServerStateRunning:
		return "Game server is running"
	case boilerrv1alpha1.ServerStateError:
		return "An error occurred"
	default:
		return ""
	}
}

// findSteamServersForGameDef returns reconcile requests for all SteamServers that reference a GameDefinition.
func (r *SteamServerReconciler) findSteamServersForGameDef(ctx context.Context, obj client.Object) []reconcile.Request {
	gameDef := obj.(*boilerrv1alpha1.GameDefinition)

	var serverList boilerrv1alpha1.SteamServerList
	if err := r.List(ctx, &serverList); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, server := range serverList.Items {
		if server.Spec.GameDefinition == gameDef.Name {
			requests = append(requests, reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&server),
			})
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *SteamServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&boilerrv1alpha1.SteamServer{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&boilerrv1alpha1.GameDefinition{},
			handler.EnqueueRequestsFromMapFunc(r.findSteamServersForGameDef),
		).
		Named("steamserver").
		Complete(r)
}

// commonLabels returns the common labels for managed resources.
func commonLabels(name, gameDefinition string) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":       "steamserver",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/managed-by": "boilerr",
	}
	if gameDefinition != "" {
		labels["boilerr.dev/game"] = gameDefinition
	}
	return labels
}

// hasLabels checks if actual labels contain all expected labels.
func hasLabels(actual, expected map[string]string) bool {
	for k, v := range expected {
		if actual[k] != v {
			return false
		}
	}
	return true
}

// portsEqual compares two PortStatus slices.
func portsEqual(a, b []boilerrv1alpha1.PortStatus) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Port != b[i].Port || a[i].Protocol != b[i].Protocol {
			return false
		}
	}
	return true
}
