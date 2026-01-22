// internal/k8s/resources.go
package k8s

import (
	"context"
	"fmt"

	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListSecrets returns all secrets in a namespace or all namespaces
func (c *Client) ListSecrets(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	var secrets []models.KubernetesResource

	if namespace == "" {
		namespaces, err := c.ListNamespaces(ctx)
		if err != nil {
			return nil, err
		}

		for _, ns := range namespaces {
			nsSecrets, err := c.listSecretsInNamespace(ctx, ns)
			if err != nil {
				continue // Skip namespaces we can't access
			}
			secrets = append(secrets, nsSecrets...)
		}
	} else {
		return c.listSecretsInNamespace(ctx, namespace)
	}

	return secrets, nil
}

func (c *Client) listSecretsInNamespace(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	secretList, err := c.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var secrets []models.KubernetesResource
	for _, secret := range secretList.Items {
		secrets = append(secrets, convertSecretToModel(&secret))
	}

	return secrets, nil
}

// ListConfigMaps returns all configmaps in a namespace or all namespaces
func (c *Client) ListConfigMaps(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	var configMaps []models.KubernetesResource

	if namespace == "" {
		namespaces, err := c.ListNamespaces(ctx)
		if err != nil {
			return nil, err
		}

		for _, ns := range namespaces {
			nsCMs, err := c.listConfigMapsInNamespace(ctx, ns)
			if err != nil {
				continue
			}
			configMaps = append(configMaps, nsCMs...)
		}
	} else {
		return c.listConfigMapsInNamespace(ctx, namespace)
	}

	return configMaps, nil
}

func (c *Client) listConfigMapsInNamespace(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	cmList, err := c.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var configMaps []models.KubernetesResource
	for _, cm := range cmList.Items {
		configMaps = append(configMaps, convertConfigMapToModel(&cm))
	}

	return configMaps, nil
}

// ListPods returns all pods in a namespace or all namespaces
func (c *Client) ListPods(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	var pods []models.KubernetesResource

	if namespace == "" {
		namespaces, err := c.ListNamespaces(ctx)
		if err != nil {
			return nil, err
		}

		for _, ns := range namespaces {
			nsPods, err := c.listPodsInNamespace(ctx, ns)
			if err != nil {
				continue
			}
			pods = append(pods, nsPods...)
		}
	} else {
		return c.listPodsInNamespace(ctx, namespace)
	}

	return pods, nil
}

func (c *Client) listPodsInNamespace(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	podList, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var pods []models.KubernetesResource
	for _, pod := range podList.Items {
		pods = append(pods, convertPodToModel(&pod))
	}

	return pods, nil
}

// ListServices returns all services in a namespace or all namespaces
func (c *Client) ListServices(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	var services []models.KubernetesResource

	if namespace == "" {
		namespaces, err := c.ListNamespaces(ctx)
		if err != nil {
			return nil, err
		}

		for _, ns := range namespaces {
			nsSvcs, err := c.listServicesInNamespace(ctx, ns)
			if err != nil {
				continue
			}
			services = append(services, nsSvcs...)
		}
	} else {
		return c.listServicesInNamespace(ctx, namespace)
	}

	return services, nil
}

func (c *Client) listServicesInNamespace(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	svcList, err := c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var services []models.KubernetesResource
	for _, svc := range svcList.Items {
		services = append(services, convertServiceToModel(&svc))
	}

	return services, nil
}

// ListKubernetesResources lists resources of a given type
func (c *Client) ListKubernetesResources(ctx context.Context, resourceType string, namespace string) ([]models.KubernetesResource, error) {
	switch resourceType {
	case "secrets", "secret":
		return c.ListSecrets(ctx, namespace)
	case "configmaps", "configmap":
		return c.ListConfigMaps(ctx, namespace)
	case "pods", "pod":
		return c.ListPods(ctx, namespace)
	case "services", "service":
		return c.ListServices(ctx, namespace)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// Converter functions
func convertSecretToModel(secret *corev1.Secret) models.KubernetesResource {
	return models.KubernetesResource{
		Kind:        "Secret",
		Name:        secret.Name,
		Namespace:   secret.Namespace,
		APIVersion:  secret.APIVersion,
		CreatedAt:   secret.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		Labels:      secret.Labels,
		Annotations: secret.Annotations,
	}
}

func convertConfigMapToModel(cm *corev1.ConfigMap) models.KubernetesResource {
	return models.KubernetesResource{
		Kind:        "ConfigMap",
		Name:        cm.Name,
		Namespace:   cm.Namespace,
		APIVersion:  cm.APIVersion,
		CreatedAt:   cm.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		Labels:      cm.Labels,
		Annotations: cm.Annotations,
	}
}

func convertPodToModel(pod *corev1.Pod) models.KubernetesResource {
	return models.KubernetesResource{
		Kind:        "Pod",
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		APIVersion:  pod.APIVersion,
		CreatedAt:   pod.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		Labels:      pod.Labels,
		Annotations: pod.Annotations,
	}
}

func convertServiceToModel(svc *corev1.Service) models.KubernetesResource {
	return models.KubernetesResource{
		Kind:        "Service",
		Name:        svc.Name,
		Namespace:   svc.Namespace,
		APIVersion:  svc.APIVersion,
		CreatedAt:   svc.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		Labels:      svc.Labels,
		Annotations: svc.Annotations,
	}
}
