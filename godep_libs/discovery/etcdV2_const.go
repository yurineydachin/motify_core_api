package discovery

// This constants are moved from discovery/discovery package.
// TODO: purge after dropping etcd2 support

const (
	// NodesProperty used to discovery nodes
	NodesProperty = "nodes"
	// VersionsProperty used to discovery nodes versions
	VersionsProperty = "versions"
	// MetricsProperty used to discovery metrics for nodes
	MetricsProperty = "metrics"
)
