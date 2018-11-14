package istio

import (
	"context"

	v12 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type EncryptionSyncer struct {
}

// TODO: This is boilerplate / duped
func (s *EncryptionSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	for _, mesh := range snap.Meshes.List() {
		switch mesh.TargetMesh.MeshType {
		case v1.MeshType_ISTIO:
			encryption := mesh.Encryption
			if encryption == nil {
				return nil
			}
			encryptionSecret := encryption.Secret
			if encryptionSecret == nil {
				return nil
			}
			secret, err := snap.Secrets.List().Find(encryptionSecret.Namespace, encryptionSecret.Name)
			if err != nil {
				return err
			}
			tlsSecret := secret.GetTls()
			if tlsSecret == nil {
				return errors.Errorf("missing tls secret")
			}

			s.sync(ctx, tlsSecret)
		}
	}
	return nil
}

//func getKubeClient() *kubernetes.Clientset {
//	kubeconfig := filepath.Join(
//		os.Getenv("HOME"), ".kube", "config",
//	)
//	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
//	if err != nil {
//		panic(err)
//	}
//
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		panic(err)
//	}
//	return clientset
//}

func (s *EncryptionSyncer) sync(ctx context.Context, secret *v12.TlsSecret) error {
	//// kubectl create secret generic cacerts -n istio-system --from-file %s --from-file %s --from-file %s --from-file %s
	//
	//d
	//
	//newCacerts = corev1.Secret{
	//	Type: corev1.SecretTypeOpaque,
	//	Data:
	//}
	//
	//client := getKubeClient()
	//getOpts := v13.GetOptions{}
	//cacerts, err := client.CoreV1().Secrets("istio-system").Get("cacerts", getOpts)
	//if err != nil || cacerts == nil {
	//	// doesn't already exist, create
	//
	//	client.CoreV1().Secrets("istio-system").Create()
	//}

}
