package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// InstallDashboard installs the Kubernetes dashboard using helm chart
func InstallDashboard(host, user, key, hostname string) (string, error) {
	cmd := "helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to add helm repo: %v\nOutput: %s", err, out)
	}
	cmd = "helm repo update"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to update helm repo: %v\nOutput: %s", err, out)
	}
	cmd = "helm install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --namespace kubernetes-dashboard --create-namespace"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to install dashboard: %v\nOutput: %s", err, out)
	}
	cmd = "kubectl get pods -n kubernetes-dashboard"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to get dashboard pods: %v\nOutput: %s", err, out)
	}
	// create dashboard ingress
	ingressYaml := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubernetes-dashboard
  namespace: kubernetes-dashboard
spec:
  rules:
    - host: %s
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kubernetes-dashboard-web
                port:
                  number: 8000`, hostname)
	log.Println(ingressYaml)
	cmd = fmt.Sprintf("cat << 'EOF' > /tmp/kubernetes-dashboard-ingress.yaml\n%s\nEOF", ingressYaml)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to create dashboard ingress: %v\nOutput: %s", err, out)
	}
	cmd = "kubectl apply -f /tmp/kubernetes-dashboard-ingress.yaml -n kubernetes-dashboard --wait=true --timeout=300s"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to apply dashboard ingress: %v\nOutput: %s", err, out)
	}
	// get dashboard token
	saYaml := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: kubernetes-dashboard`
	cmd = fmt.Sprintf("cat << 'EOF' > /tmp/kubernetes-dashboard-sa.yaml\n%s\nEOF", saYaml)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to create dashboard service account: %v\nOutput: %s", err, out)
	}
	cmd = "kubectl apply -f /tmp/kubernetes-dashboard-sa.yaml -n kubernetes-dashboard --wait=true --timeout=300s"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to apply dashboard service account: %v\nOutput: %s", err, out)
	}
	crBindingRole := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: admin-user
    namespace: kubernetes-dashboard
`
	cmd = fmt.Sprintf("cat << 'EOF' > /tmp/kubernetes-dashboard-crb.yaml\n%s\nEOF", crBindingRole)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to create dashboard cluster role binding: %v\nOutput: %s", err, out)
	}
	cmd = "kubectl apply -f /tmp/kubernetes-dashboard-crb.yaml -n kubernetes-dashboard --wait=true --timeout=300s"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to apply dashboard cluster role binding: %v\nOutput: %s", err, out)
	}
	tokenSecret := `apiVersion: v1
kind: Secret
metadata:
  name: admin-user
  namespace: kubernetes-dashboard
  annotations:
    kubernetes.io/service-account.name: "admin-user"   
type: kubernetes.io/service-account-token`
	cmd = fmt.Sprintf("cat << 'EOF' > /tmp/kubernetes-dashboard-token-secret.yaml\n%s\nEOF", tokenSecret)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to create dashboard token secret: %v\nOutput: %s", err, out)
	}
	cmd = "kubectl apply -f /tmp/kubernetes-dashboard-token-secret.yaml -n kubernetes-dashboard --wait=true --timeout=300s"
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 600); err != nil {
		return "", fmt.Errorf("failed to apply dashboard token secret: %v\nOutput: %s", err, out)
	}
	cmd = "kubectl -n kubernetes-dashboard get secret admin-user -o jsonpath='{.data.token}'"
	log.Println(cmd)
	out, err := ssh.Run(cmd, host, key, user, "", true, 600)
	if err != nil {
		return "", fmt.Errorf("failed to get dashboard token: %v\nOutput: %s", err, out)
	}
	return out, nil
}
