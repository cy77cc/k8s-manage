package cluster

import "time"

const (
	clusterCachePrefix = "cluster"

	CacheTTLShort  = 20 * time.Second
	CacheTTLMedium = 60 * time.Second
	CacheTTLLong   = 2 * time.Minute
)

type CachePolicy struct {
	KeyPattern       string
	TTL              time.Duration
	InvalidatesOnOps []string
}

// ClusterPhase1CachePolicies defines governed key namespaces and TTL classes.
var ClusterPhase1CachePolicies = map[string]CachePolicy{
	"clusters.list": {
		KeyPattern:       "cluster:list:{status}:{source}",
		TTL:              CacheTTLShort,
		InvalidatesOnOps: []string{"cluster.create", "cluster.update", "cluster.delete", "cluster.import", "cluster.sync"},
	},
	"clusters.detail": {
		KeyPattern:       "cluster:detail:{id}",
		TTL:              CacheTTLMedium,
		InvalidatesOnOps: []string{"cluster.update", "cluster.delete", "cluster.import", "cluster.sync"},
	},
	"clusters.nodes": {
		KeyPattern:       "cluster:nodes:{id}",
		TTL:              CacheTTLShort,
		InvalidatesOnOps: []string{"cluster.addNode", "cluster.removeNode", "cluster.sync", "cluster.import"},
	},
	"clusters.bootstrap_profiles": {
		KeyPattern:       "cluster:bootstrap:profiles:list",
		TTL:              CacheTTLLong,
		InvalidatesOnOps: []string{"cluster.bootstrap_profile.create", "cluster.bootstrap_profile.update", "cluster.bootstrap_profile.delete"},
	},
}

func CacheKeyClusterList(status, source string) string {
	return clusterCachePrefix + ":list:" + status + ":" + source
}

func CacheKeyClusterDetail(id uint) string {
	return clusterCachePrefix + ":detail:" + itoa(id)
}

func CacheKeyClusterNodes(id uint) string {
	return clusterCachePrefix + ":nodes:" + itoa(id)
}

func CacheKeyBootstrapProfiles() string {
	return clusterCachePrefix + ":bootstrap:profiles:list"
}

func itoa(v uint) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	return string(buf[i:])
}
