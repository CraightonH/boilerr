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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

// GameDefinitionReconciler reconciles a GameDefinition object.
type GameDefinitionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=boilerr.dev,resources=gamedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=boilerr.dev,resources=gamedefinitions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=boilerr.dev,resources=gamedefinitions/finalizers,verbs=update

// Reconcile validates and updates the status of a GameDefinition.
func (r *GameDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch GameDefinition
	var gameDef boilerrv1alpha1.GameDefinition
	if err := r.Get(ctx, req.NamespacedName, &gameDef); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate the GameDefinition
	if err := r.validate(&gameDef); err != nil {
		gameDef.Status.Ready = false
		gameDef.Status.Message = err.Error()
		if statusErr := r.Status().Update(ctx, &gameDef); statusErr != nil {
			logger.Error(statusErr, "Failed to update GameDefinition status")
			return ctrl.Result{}, statusErr
		}
		// Don't requeue - user needs to fix the definition
		return ctrl.Result{}, nil
	}

	// Mark ready
	if !gameDef.Status.Ready || gameDef.Status.Message != "GameDefinition validated successfully" {
		gameDef.Status.Ready = true
		gameDef.Status.Message = "GameDefinition validated successfully"
		if err := r.Status().Update(ctx, &gameDef); err != nil {
			logger.Error(err, "Failed to update GameDefinition status")
			return ctrl.Result{}, err
		}
		logger.Info("GameDefinition validated", "name", gameDef.Name)
	}

	return ctrl.Result{}, nil
}

// validate checks that the GameDefinition has all required fields.
func (r *GameDefinitionReconciler) validate(gd *boilerrv1alpha1.GameDefinition) error {
	if gd.Spec.AppId <= 0 {
		return fmt.Errorf("appId must be a positive integer")
	}
	if gd.Spec.Command == "" {
		return fmt.Errorf("command is required")
	}
	if len(gd.Spec.Ports) == 0 {
		return fmt.Errorf("at least one port is required")
	}

	// Validate ports have required fields
	for i, port := range gd.Spec.Ports {
		if port.Name == "" {
			return fmt.Errorf("port[%d].name is required", i)
		}
		if port.ContainerPort <= 0 || port.ContainerPort > 65535 {
			return fmt.Errorf("port[%d].containerPort must be between 1 and 65535", i)
		}
	}

	// Validate configSchema entries
	for key, entry := range gd.Spec.ConfigSchema {
		if entry.MapTo != nil {
			switch entry.MapTo.Type {
			case "arg", "env", "configFile":
				// valid
			default:
				return fmt.Errorf("configSchema[%s].mapTo.type must be 'arg', 'env', or 'configFile'", key)
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&boilerrv1alpha1.GameDefinition{}).
		Named("gamedefinition").
		Complete(r)
}
