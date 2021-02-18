/*
Copyright 2021.

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

package v1alpha1

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var aerospikeclusterlog = logf.Log.WithName("aerospikecluster-resource")

func (r *AerospikeCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// // EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// // +kubebuilder:webhook:path=/mutate-aerospike-aerospike-com-v1alpha1-aerospikecluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=aerospike.aerospike.com,resources=aerospikeclusters,verbs=create;update,versions=v1alpha1,name=maerospikecluster.kb.io,admissionReviewVersions={v1,v1beta1}

// var _ webhook.Defaulter = &AerospikeCluster{}

// // Default implements webhook.Defaulter so a webhook will be registered for the type
// func (r *AerospikeCluster) Default() {
// 	aerospikeclusterlog.Info("default", "name", r.Name)

// 	// TODO(user): fill in your defaulting logic.
// }

// // TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// // +kubebuilder:webhook:path=/validate-aerospike-aerospike-com-v1alpha1-aerospikecluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=aerospike.aerospike.com,resources=aerospikeclusters,verbs=create;update,versions=v1alpha1,name=vaerospikecluster.kb.io,admissionReviewVersions={v1,v1beta1}

// var _ webhook.Validator = &AerospikeCluster{}

// // ValidateCreate implements webhook.Validator so a webhook will be registered for the type
// func (r *AerospikeCluster) ValidateCreate() error {
// 	aerospikeclusterlog.Info("validate create", "name", r.Name)

// 	// TODO(user): fill in your validation logic upon object creation.
// 	return nil
// }

// // ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
// func (r *AerospikeCluster) ValidateUpdate(old runtime.Object) error {
// 	aerospikeclusterlog.Info("validate update", "name", r.Name)

// 	// TODO(user): fill in your validation logic upon object update.
// 	return nil
// }

// // ValidateDelete implements webhook.Validator so a webhook will be registered for the type
// func (r *AerospikeCluster) ValidateDelete() error {
// 	aerospikeclusterlog.Info("validate delete", "name", r.Name)

// 	// TODO(user): fill in your validation logic upon object deletion.
// 	return nil
// }
