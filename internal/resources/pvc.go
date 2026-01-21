package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

// PVCBuilder builds a PersistentVolumeClaim for a SteamServer.
type PVCBuilder struct {
	server *boilerrv1alpha1.SteamServer
}

// NewPVCBuilder creates a new PVCBuilder.
func NewPVCBuilder(server *boilerrv1alpha1.SteamServer) *PVCBuilder {
	return &PVCBuilder{server: server}
}

// Build creates the PVC for the SteamServer.
func (b *PVCBuilder) Build() *corev1.PersistentVolumeClaim {
	labels := b.labels()

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PVCName(b.server.Name),
			Namespace: b.server.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: b.server.Spec.Storage.Size,
				},
			},
		},
	}

	// Set storage class if specified
	if b.server.Spec.Storage.StorageClassName != nil {
		pvc.Spec.StorageClassName = b.server.Spec.Storage.StorageClassName
	}

	return pvc
}

// labels returns the common labels for the PVC.
func (b *PVCBuilder) labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "steamserver",
		"app.kubernetes.io/instance":   b.server.Name,
		"app.kubernetes.io/managed-by": "boilerr",
	}
}
