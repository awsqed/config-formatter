package traefik

import (
	"sort"
	"strings"

	"github.com/awsqed/config-formatter/formatter"
	"gopkg.in/yaml.v3"
)

// TraefikFormatter formats Traefik configuration files
type TraefikFormatter struct {
	formatter.BaseFormatter
}

// New creates a new TraefikFormatter
func New() *TraefikFormatter {
	return &TraefikFormatter{}
}

// Name returns the name of this formatter
func (f *TraefikFormatter) Name() string {
	return "traefik"
}

// CanHandle checks if this file is a Traefik configuration file
func (f *TraefikFormatter) CanHandle(filename string, data []byte) bool {
	// Check filename patterns
	if strings.Contains(filename, "traefik") {
		return true
	}

	// Check for traefik-specific keys
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return false
	}

	// Look for traefik indicators in top-level keys
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		content := root.Content[0]
		if content.Kind == yaml.MappingNode {
			for i := 0; i < len(content.Content); i += 2 {
				key := content.Content[i].Value
				if key == "http" || key == "tcp" || key == "udp" ||
					key == "entryPoints" || key == "providers" ||
					key == "certificatesResolvers" || key == "api" {
					return true
				}
			}
		}
	}

	return false
}

// Format formats a Traefik YAML file with consistent indentation and ordering
func (f *TraefikFormatter) Format(data []byte, indent int) ([]byte, error) {
	return f.FormatYAML(data, indent, f.formatNode)
}

// formatNode recursively formats nodes in the YAML tree
func (f *TraefikFormatter) formatNode(node *yaml.Node, isRoot bool) {
	if node == nil {
		return
	}

	// Process mapping nodes (objects)
	if node.Kind == yaml.MappingNode {
		f.sortMappingNode(node, isRoot)
	}

	// Recursively format child nodes
	// Check if this is the root document node
	if isRoot && node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		f.formatNode(node.Content[0], true)
		return
	}

	for _, child := range node.Content {
		f.formatNode(child, false)
	}
}

// sortMappingNode sorts keys in a mapping node according to Traefik conventions
func (f *TraefikFormatter) sortMappingNode(node *yaml.Node, isTopLevel bool) {
	if node.Kind != yaml.MappingNode || len(node.Content) == 0 {
		return
	}

	// Create pairs of key-value nodes
	type pair struct {
		key         *yaml.Node
		value       *yaml.Node
		order       int
		originalIdx int
		hasComment  bool
	}

	var pairs []pair

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		hasComment := keyNode.HeadComment != "" || keyNode.LineComment != "" ||
			keyNode.FootComment != "" || valueNode.HeadComment != ""

		pairs = append(pairs, pair{
			key:         keyNode,
			value:       valueNode,
			order:       getKeyOrder(keyNode.Value, isTopLevel),
			originalIdx: i,
			hasComment:  hasComment,
		})
	}

	// Sort pairs by order, then alphabetically, but keep commented blocks in original position
	sort.SliceStable(pairs, func(i, j int) bool {
		// If either pair has comments, preserve original order relative to each other
		if pairs[i].hasComment || pairs[j].hasComment {
			return pairs[i].originalIdx < pairs[j].originalIdx
		}

		if pairs[i].order != pairs[j].order {
			return pairs[i].order < pairs[j].order
		}
		return pairs[i].key.Value < pairs[j].key.Value
	})

	// Add empty lines between top-level directives AFTER sorting
	if isTopLevel {
		for i := 1; i < len(pairs); i++ {
			keyNode := pairs[i].key
			if keyNode.HeadComment == "" {
				keyNode.HeadComment = "\n"
			} else if keyNode.HeadComment[0] != '\n' {
				keyNode.HeadComment = "\n" + keyNode.HeadComment
			}
		}
	}

	// Rebuild the Content slice with sorted pairs
	newContent := make([]*yaml.Node, 0, len(node.Content))
	for _, p := range pairs {
		newContent = append(newContent, p.key, p.value)
	}
	node.Content = newContent
}

// getKeyOrder returns the sort order for Traefik configuration keys
// Lower numbers come first
func getKeyOrder(key string, isTopLevel bool) int {
	// Top-level Traefik configuration keys order
	topLevelOrder := map[string]int{
		// Global configuration
		"global":     1,
		"log":        2,
		"accessLog":  3,
		"api":        4,
		"ping":       5,
		"metrics":    6,
		"tracing":    7,
		"hostResolver": 8,

		// Entry points and providers
		"entryPoints":           10,
		"providers":             11,
		"certificatesResolvers": 12,

		// Protocol-specific configurations (http, tcp, udp)
		"http": 100,
		"tcp":  101,
		"udp":  102,
		"tls":  103,

		// Experimental and other
		"experimental": 200,
		"pilot":        201,
	}

	// HTTP section keys order (http.routers, http.services, etc.)
	httpOrder := map[string]int{
		"routers":           1,
		"services":          2,
		"middlewares":       3,
		"models":            4,
		"serversTransports": 5,
	}

	// TCP section keys order
	tcpOrder := map[string]int{
		"routers":           1,
		"services":          2,
		"middlewares":       3,
		"serversTransports": 4,
	}

	// UDP section keys order
	udpOrder := map[string]int{
		"routers":  1,
		"services": 2,
	}

	// Router configuration keys order (applies to http/tcp routers)
	routerOrder := map[string]int{
		"entryPoints": 1,
		"rule":        2,
		"priority":    3,
		"service":     4,
		"middlewares": 5,
		"tls":         6,
	}

	// Service configuration keys order
	serviceOrder := map[string]int{
		"loadBalancer": 1,
		"weighted":     2,
		"mirroring":    3,
		"failover":     4,
	}

	// LoadBalancer configuration keys order
	loadBalancerOrder := map[string]int{
		"servers":            1,
		"healthCheck":        2,
		"sticky":             3,
		"serversTransport":   4,
		"passHostHeader":     5,
		"responseForwarding": 6,
	}

	// Middleware configuration keys order
	middlewareOrder := map[string]int{
		"addPrefix":          1,
		"stripPrefix":        2,
		"stripPrefixRegex":   3,
		"replacePath":        4,
		"replacePathRegex":   5,
		"chain":              6,
		"ipWhiteList":        7,
		"ipAllowList":        8,
		"headers":            9,
		"errors":             10,
		"rateLimit":          11,
		"redirectRegex":      12,
		"redirectScheme":     13,
		"basicAuth":          14,
		"digestAuth":         15,
		"forwardAuth":        16,
		"inFlightReq":        17,
		"buffering":          18,
		"circuitBreaker":     19,
		"compress":           20,
		"contentType":        21,
		"grpcWeb":            22,
		"passTLSClientCert":  23,
		"retry":              24,
	}

	// Entry point configuration keys order
	entryPointOrder := map[string]int{
		"address":          1,
		"asDefault":        2,
		"transport":        3,
		"http":             4,
		"http2":            5,
		"http3":            6,
		"proxyProtocol":    7,
		"forwardedHeaders": 8,
	}

	// TLS configuration keys order
	tlsOrder := map[string]int{
		"certificates": 1,
		"options":      2,
		"stores":       3,
		"certResolver": 4,
		"domains":      5,
	}

	// Provider configuration keys order
	providerOrder := map[string]int{
		"providersThrottleDuration": 1,
		"docker":                    10,
		"file":                      11,
		"kubernetes":                12,
		"kubernetesGateway":         13,
		"kubernetesCRD":             14,
		"consulCatalog":             15,
		"nomad":                     16,
		"ecs":                       17,
		"marathon":                  18,
		"rancher":                   19,
		"rest":                      20,
		"etcd":                      21,
		"consul":                    22,
		"zooKeeper":                 23,
		"redis":                     24,
		"http":                      25,
	}

	// Check which order map to use based on common patterns
	if order, ok := topLevelOrder[key]; ok {
		return order
	}
	if order, ok := httpOrder[key]; ok {
		return order
	}
	if order, ok := tcpOrder[key]; ok {
		return order
	}
	if order, ok := udpOrder[key]; ok {
		return order
	}
	if order, ok := routerOrder[key]; ok {
		return order
	}
	if order, ok := serviceOrder[key]; ok {
		return order
	}
	if order, ok := loadBalancerOrder[key]; ok {
		return order
	}
	if order, ok := middlewareOrder[key]; ok {
		return order
	}
	if order, ok := entryPointOrder[key]; ok {
		return order
	}
	if order, ok := tlsOrder[key]; ok {
		return order
	}
	if order, ok := providerOrder[key]; ok {
		return order
	}

	// Default order for unknown keys (alphabetical sorting will apply)
	return 1000
}
