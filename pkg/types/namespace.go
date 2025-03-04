package types

// Namespace is a namespace
type Namespace struct {
	Name string `json:"name"`
}

// ReservedNamespaces is a map of reserved namespaces
var ReservedNamespaces = map[string]bool{
	"default":                true,
	"longhorn-system":        true,
	"kube-system":            true,
	"kube-public":            true,
	"kube-node-lease":        true,
	"cnpg-system":            true,
	"clickhouse-system":      true,
	"kubevirt-manager":       true,
	"kubevirt":               true,
	"kubernetes-dashboard":   true,
	"monitoring":             true,
	"mysql-operator":         true,
	"ollama-operator-system": true,
}
