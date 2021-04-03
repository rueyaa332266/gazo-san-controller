/*


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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gazosancontrollerv1alpha1 "github.com/rueyaa332266/gazo-san-controller/api/v1alpha1"
)

// ReportReconciler reconciles a Report object
type ReportReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=gazosancontroller.k8s.io,resources=reports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gazosancontroller.k8s.io,resources=reports/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=developments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ReportReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("report", req.NamespacedName)

	// Load the Report by name
	var report gazosancontrollerv1alpha1.Report
	log.Info("fetching Report Resource")
	if err := r.Get(ctx, req.NamespacedName, &report); err != nil {
		log.Error(err, "unable to fetch Report")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create or Update deployment object which match report.Spec.
	deploymentName := "gazo-san-report"
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: req.Namespace,
		},
	}

	// Create or Update deployment object
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		baseURL := report.Spec.BaseURL
		compareURL := report.Spec.CompareURL
		replicas := int32(1)
		deploy.Spec.Replicas = &replicas

		// set a label for our deployment
		labels := map[string]string{
			"app":        "gazo-san-report",
			"controller": req.Name,
		}

		// set labels to spec.selector for our deployment
		if deploy.Spec.Selector == nil {
			deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		// set labels to template.objectMeta for our deployment
		if deploy.Spec.Template.ObjectMeta.Labels == nil {
			deploy.Spec.Template.ObjectMeta.Labels = labels
		}

		// set a container for our deployment
		containers := []corev1.Container{
			{
				Name:  "gazo-san-report",
				Image: "aa332266/gazo-san-report:latest",
				Env: []corev1.EnvVar{
					{
						Name:  "BaseURL",
						Value: baseURL,
					}, {
						Name:  "CompareURL",
						Value: compareURL,
					},
				},
				Args: []string{"local-server"},
			},
		}

		// set containers to template.spec.containers for our deployment
		if deploy.Spec.Template.Spec.Containers == nil {
			deploy.Spec.Template.Spec.Containers = containers
		}

		// set the owner so that garbage collection can kicks in
		if err := ctrl.SetControllerReference(&report, deploy, r.Scheme); err != nil {
			log.Error(err, "unable to set ownerReference from Report to Deployment")
			return err
		}

		// end of ctrl.CreateOrUpdate
		return nil
	}); err != nil {

		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure deployment is correct")
		return ctrl.Result{}, err

	}

	return ctrl.Result{}, nil
}

var (
	deploymentOwnerKey = ".metadata.controller"
	apiGVStr           = gazosancontrollerv1alpha1.GroupVersion.String()
)

func (r *ReportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, deploymentOwnerKey, func(rawObj runtime.Object) []string {
		// grab the deployment object, extract the owner...
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Report...
		if owner.APIVersion != apiGVStr || owner.Kind != "Report" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&gazosancontrollerv1alpha1.Report{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
