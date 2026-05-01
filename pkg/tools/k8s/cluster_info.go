package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getK8sClusterInfo(ctx context.Context, _ *mcp.CallToolRequest, args *getK8sClusterInfoArgs) (*mcp.CallToolResult, any, error) {
	clientset, res, err := getK8sClient(&args.Cluster)
	if res != nil || err != nil {
		return res, nil, err
	}

	clusterInfo, err := getK8sClusterInfoImpl(ctx, clientset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return textResult(clusterInfo), nil, nil
}

func getK8sClusterInfoImpl(ctx context.Context, client *Clientset) (string, error) {
	services, err := listServices(ctx, client.Client, "kube-system", &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"kubernetes.io/cluster-service": "true",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to list services %w", err)
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("Kubernetes control plane is running at %s", client.RestConfig.Host))
	for _, s := range services {
		lines = append(lines, fmt.Sprintf("%s is running at %s", serviceName(s), serviceLink(s, client.RestConfig, client.Client)))
	}
	return strings.Join(lines, "\n"), nil
}

func listServices(ctx context.Context, client kubernetes.Interface, namespace string, labelSelector *metav1.LabelSelector) ([]corev1.Service, error) {
	services, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(labelSelector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services %w", err)
	}
	return services.Items, nil
}

func serviceName(s corev1.Service) string {
	if name, found := s.ObjectMeta.GetLabels()["kubernetes.io/name"]; found {
		return name
	}
	return s.ObjectMeta.GetName()
}

func serviceLink(s corev1.Service, rc *rest.Config, client kubernetes.Interface) string {
	if len(s.Status.LoadBalancer.Ingress) > 0 {
		ingress := s.Status.LoadBalancer.Ingress[0]
		ip := ingress.IP
		if ip == "" {
			ip = ingress.Hostname
		}
		var link strings.Builder
		for _, port := range s.Spec.Ports {
			link.WriteString("http://" + ip + ":" + strconv.Itoa(int(port.Port)))
		}
		return link.String()
	}
	return rc.Host + "/api/v1/namespaces/" + s.ObjectMeta.Namespace + "/services/" + serviceProxyResourceName(s) + "/proxy"
}

func serviceProxyResourceName(service corev1.Service) string {
	serviceName := service.ObjectMeta.Name
	if len(service.Spec.Ports) > 0 {
		port := service.Spec.Ports[0]
		if scheme := guessScheme(port); len(scheme) > 0 {
			return scheme + ":" + serviceName + ":" + port.Name
		}
		if len(port.Name) > 0 {
			return serviceName + ":" + port.Name
		}
		return serviceName
	}
	return service.ObjectMeta.Name
}

func guessScheme(port corev1.ServicePort) string {
	if port.Name == "https" || port.Port == 443 {
		// Use same heuristic as kubectl to "guess" if the service is using HTTPS.
		return "https"
	}
	return ""
}
