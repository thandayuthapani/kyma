package resources

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CertFile       = "server-cert.pem"
	KeyFile        = "server-key.pem"
	DefaultCertDir = "/tmp/k8s-webhook-server/serving-certs"
)

func SetupCertificates(ctx context.Context, secretName, secretNamespace, serviceName string) error {
	// We are going to talk to the API server _before_ we start the manager.
	// Since the default manager client reads from cache, we will get an error.
	// So, we create a "serverClient" that would read from the API directly.
	// We only use it here, this only runs at start up, so it shouldn't be to much for the API
	serverClient, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{})
	if err != nil {
		return errors.Wrap(err, "failed to create a server client")
	}
	if err := EnsureWebhookSecret(ctx, serverClient, secretName, secretNamespace, serviceName); err != nil {
		return errors.Wrap(err, "failed to ensure webhook secret")
	}
	return nil
}

func EnsureWebhookSecret(ctx context.Context, client ctrlclient.Client, secretName, secretNamespace, serviceName string) error {
	logger := ctrl.LoggerFrom(ctx)
	secret := &corev1.Secret{}
	logger.Info("ensuring webhook secret")
	err := client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to get webhook secret")
	}

	if apiErrors.IsNotFound(err) {
		logger.Info("creating webhook secret")
		return createSecret(ctx, client, secretName, secretNamespace, serviceName)
	}

	logger.Info("updating pre-exiting webhook secret")
	if err := updateSecret(ctx, client, logger, secret, serviceName); err != nil {
		return errors.Wrap(err, "failed to update secret")
	}
	return nil
}

func createSecret(ctx context.Context, client ctrlclient.Client, name, namespace, serviceName string) error {
	secret, err := buildSecret(name, namespace, serviceName)
	if err != nil {
		return errors.Wrap(err, "failed to create secret object")
	}
	if err := client.Create(ctx, secret); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}
	return nil
}

func updateSecret(ctx context.Context, client ctrlclient.Client, log logr.Logger, secret *corev1.Secret, serviceName string) error {
	valid, err := isValidSecret(secret)
	if valid {
		return nil
	}
	if err != nil {
		log.Error(err, "invalid certificate")
	}

	newSecret, err := buildSecret(secret.Name, secret.Namespace, serviceName)
	if err != nil {
		return errors.Wrap(err, "failed to create secret object")
	}

	secret.Data = newSecret.Data
	if err := client.Update(ctx, secret); err != nil {
		return errors.Wrap(err, "failed to update secret")
	}
	return nil
}

func isValidSecret(s *corev1.Secret) (bool, error) {
	if !hasRequiredKeys(s.Data) {
		return false, nil
	}
	if err := verifyCertificate(s.Data[CertFile]); err != nil {
		return false, err
	}
	if err := verifyKey(s.Data[KeyFile]); err != nil {
		return false, err
	}

	return true, nil
}

func verifyCertificate(c []byte) error {
	certificate, err := cert.ParseCertsPEM(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse certificate data")
	}
	// certificate is self signed. So we use it as a root cert
	root, err := cert.NewPoolFromBytes(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse root certificate data")
	}
	// make sure the certificate is valid for the next 10 days. Otherwise it will be recreated.
	_, err = certificate[0].Verify(x509.VerifyOptions{CurrentTime: time.Now().Add(10 * 24 * time.Hour), Roots: root})
	if err != nil {
		return errors.Wrap(err, "certificate verification failed")
	}
	return nil
}

func verifyKey(k []byte) error {
	b, _ := pem.Decode(k)
	key, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse key data")
	}
	if err = key.Validate(); err != nil {
		return errors.Wrap(err, "key verification failed")
	}
	return nil
}

func hasRequiredKeys(data map[string][]byte) bool {
	if data == nil {
		return false
	}
	for _, key := range []string{CertFile, KeyFile} {
		if _, ok := data[key]; !ok {
			return false
		}
	}
	return true
}

func buildSecret(name, namespace, serviceName string) (*corev1.Secret, error) {
	cert, key, err := generateWebhookCertificates(serviceName, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate webhook certificates")
	}
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			CertFile: cert,
			KeyFile:  key,
		},
	}, nil
}

func generateWebhookCertificates(serviceName, namespace string) ([]byte, []byte, error) {
	altNames := serviceAltNames(serviceName, namespace)
	return cert.GenerateSelfSignedCertKey(altNames[0], nil, altNames)
}

func serviceAltNames(serviceName, namespace string) []string {
	namespacedServiceName := strings.Join([]string{serviceName, namespace}, ".")
	commonName := strings.Join([]string{namespacedServiceName, "svc"}, ".")
	serviceHostname := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)

	return []string{
		commonName,
		serviceName,
		namespacedServiceName,
		serviceHostname,
	}
}
