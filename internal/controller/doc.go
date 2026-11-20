// +kubebuilder:rbac:groups=postgres-provisioner.io,resources=postgresinstances,verbs=get;list;watch
// +kubebuilder:rbac:groups=postgres-provisioner.io,resources=postgresinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

package controller
