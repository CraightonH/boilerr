package resources

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

// DefaultStorageSize is the default storage size when not specified.
const DefaultStorageSize = "20Gi"

// PVCBuilder builds a PersistentVolumeClaim for a SteamServer.
type PVCBuilder struct {
	server  *boilerrv1alpha1.SteamServer
	gameDef *boilerrv1alpha1.GameDefinition
}

// NewPVCBuilder creates a new PVCBuilder.
// gameDef can be nil for backwards compatibility (fallback mode).
func NewPVCBuilder(server *boilerrv1alpha1.SteamServer, gameDef *boilerrv1alpha1.GameDefinition) *PVCBuilder {
	return &PVCBuilder{server: server, gameDef: gameDef}
}

// Build creates the PVC for the SteamServer.
// Returns nil if no storage is configured (neither in SteamServer nor GameDefinition).
func (b *PVCBuilder) Build() *corev1.PersistentVolumeClaim {
	storageSize := b.getStorageSize()
	if storageSize.IsZero() {
		return nil
	}

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
					corev1.ResourceStorage: storageSize,
				},
			},
		},
	}

	// Set storage class if specified
	if b.server.Spec.Storage != nil && b.server.Spec.Storage.StorageClassName != nil {
		pvc.Spec.StorageClassName = b.server.Spec.Storage.StorageClassName
	}

	return pvc
}

// getStorageSize returns the storage size.
// Fallback: SteamServer.Storage.Size -> GameDefinition.DefaultStorage -> DefaultStorageSize
func (b *PVCBuilder) getStorageSize() resource.Quantity {
	// SteamServer explicit storage takes precedence
	if b.server.Spec.Storage != nil {
		return b.server.Spec.Storage.Size
	}

	// Fall back to GameDefinition default
	if b.gameDef != nil && b.gameDef.Spec.DefaultStorage != "" {
		qty, err := resource.ParseQuantity(b.gameDef.Spec.DefaultStorage)
		if err == nil {
			return qty
		}
	}

	// Use default storage size
	qty, _ := resource.ParseQuantity(DefaultStorageSize)
	return qty
}

// labels returns the common labels for the PVC.
func (b *PVCBuilder) labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "steamserver",
		"app.kubernetes.io/instance":   b.server.Name,
		"app.kubernetes.io/managed-by": "boilerr",
		"boilerr.dev/game":             b.server.Spec.Game,
	}
}
