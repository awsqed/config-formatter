package dockercompose

import (
	"sort"
	"strconv"
	"strings"

	"github.com/awsqed/config-formatter/formatter"
	"gopkg.in/yaml.v3"
)

// DockerComposeFormatter formats docker-compose files
type DockerComposeFormatter struct {
	formatter.BaseFormatter
}

// New creates a new DockerComposeFormatter
func New() *DockerComposeFormatter {
	return &DockerComposeFormatter{}
}

// Name returns the name of this formatter
func (f *DockerComposeFormatter) Name() string {
	return "docker-compose"
}

// CanHandle checks if this file is a docker-compose file
func (f *DockerComposeFormatter) CanHandle(filename string, data []byte) bool {
	// Check filename patterns
	if strings.Contains(filename, "docker-compose") ||
		strings.Contains(filename, "compose.") ||
		strings.HasSuffix(filename, "compose.yml") ||
		strings.HasSuffix(filename, "compose.yaml") {
		return true
	}

	// Check for docker-compose specific keys
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return false
	}

	// Look for docker-compose indicators in top-level keys
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		content := root.Content[0]
		if content.Kind == yaml.MappingNode {
			for i := 0; i < len(content.Content); i += 2 {
				key := content.Content[i].Value
				if key == "services" || key == "version" {
					return true
				}
			}
		}
	}

	return false
}

// Format formats a docker-compose YAML file with consistent indentation and ordering
func (f *DockerComposeFormatter) Format(data []byte, indent int) ([]byte, error) {
	return f.FormatYAML(data, indent, f.formatNode)
}

// formatNode recursively formats nodes in the YAML tree
func (f *DockerComposeFormatter) formatNode(node *yaml.Node, isRoot bool) {
	f.formatNodeWithContext(node, isRoot, "")
}

// formatNodeWithContext recursively formats nodes with parent key tracking
func (f *DockerComposeFormatter) formatNodeWithContext(node *yaml.Node, isRoot bool, parentKey string) {
	if node == nil {
		return
	}

	// Process mapping nodes (objects)
	if node.Kind == yaml.MappingNode {
		f.sortMappingNode(node, isRoot)
	}

	// Apply value normalization AFTER sorting, BEFORE recursion
	f.normalizeValues(node, parentKey)

	// Recursively format child nodes
	// Check if this is the root document node
	if isRoot && node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		f.formatNodeWithContext(node.Content[0], true, "")
		return
	}

	// For mapping nodes, track key names when recursing into values
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			f.formatNodeWithContext(valueNode, false, keyNode.Value)
		}
	} else {
		// For sequences and other nodes, don't update parent key
		for _, child := range node.Content {
			f.formatNodeWithContext(child, false, parentKey)
		}
	}
}

// sortMappingNode sorts keys in a mapping node according to docker-compose conventions
func (f *DockerComposeFormatter) sortMappingNode(node *yaml.Node, isTopLevel bool) {
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

		// Add empty line before each service entry if this is the services block
		if isTopLevel && keyNode.Value == "services" && valueNode.Kind == yaml.MappingNode {
			// Add empty line before each service entry
			addServiceSpacing(valueNode)
		}

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

// addServiceSpacing adds empty lines between service entries
func addServiceSpacing(servicesNode *yaml.Node) {
	if servicesNode.Kind != yaml.MappingNode || len(servicesNode.Content) == 0 {
		return
	}

	// Add HeadComment with newline to each service key (except the first)
	for i := 0; i < len(servicesNode.Content); i += 2 {
		keyNode := servicesNode.Content[i]

		// Add empty line before all services except the first
		if i > 0 {
			// If there's already a comment, prepend a newline
			if keyNode.HeadComment != "" {
				// Only add if it doesn't already start with newline
				if keyNode.HeadComment[0] != '\n' {
					keyNode.HeadComment = "\n" + keyNode.HeadComment
				}
			} else {
				keyNode.HeadComment = "\n"
			}
		}
	}
}

// normalizeEnvironment converts environment array to map with smart quoting
func (f *DockerComposeFormatter) normalizeEnvironment(node *yaml.Node) {
	// Only process sequence nodes (arrays)
	if node.Kind != yaml.SequenceNode {
		return
	}

	// Parse all array items as KEY=VALUE pairs
	envMap := make(map[string]string)
	var keys []string // Preserve order

	for _, item := range node.Content {
		if item.Kind != yaml.ScalarNode {
			continue // Skip non-scalar items
		}

		key, value, ok := parseEnvVar(item.Value)
		if !ok {
			continue // Skip malformed entries
		}

		if _, exists := envMap[key]; !exists {
			keys = append(keys, key)
		}
		envMap[key] = value
	}

	// Only transform if we successfully parsed at least one entry
	if len(envMap) == 0 {
		return
	}

	// Convert node to MappingNode
	node.Kind = yaml.MappingNode
	node.Tag = "!!map"
	node.Style = 0 // Default style

	// Build new Content with key-value pairs
	newContent := make([]*yaml.Node, 0, len(keys)*2)

	for _, key := range keys {
		value := envMap[key]

		// Create key node (always unquoted)
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
			Style: 0,
		}

		// Create value node with smart quoting
		valueNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: value,
		}

		// Apply smart quoting
		if shouldQuoteValue(value) {
			valueNode.Style = yaml.DoubleQuotedStyle
		} else {
			valueNode.Style = 0 // Unquoted
		}

		newContent = append(newContent, keyNode, valueNode)
	}

	node.Content = newContent
}

// normalizePorts ensures all port strings are quoted
func (f *DockerComposeFormatter) normalizePorts(node *yaml.Node) {
	// Only process sequence nodes (arrays)
	if node.Kind != yaml.SequenceNode {
		return
	}

	// Force quoting on all scalar port entries
	for _, item := range node.Content {
		if item.Kind == yaml.ScalarNode {
			// Ensure it's tagged as string and quoted
			item.Tag = "!!str"
			item.Style = yaml.DoubleQuotedStyle
		}
	}
}

// normalizeValues dispatches to specific normalizers based on parent key
func (f *DockerComposeFormatter) normalizeValues(node *yaml.Node, parentKey string) {
	switch parentKey {
	case "environment":
		f.normalizeEnvironment(node)
	case "ports":
		f.normalizePorts(node)
	// Add other cases as needed
	}
}

// parseEnvVar splits "KEY=VALUE" into key and value
// Returns (key, value, true) on success, ("", "", false) on malformed input
func parseEnvVar(envStr string) (key, value string, ok bool) {
	// Find first '=' to split key and value
	idx := strings.Index(envStr, "=")
	if idx == -1 {
		return "", "", false
	}

	key = envStr[:idx]
	value = envStr[idx+1:] // Everything after first '='

	// Key must not be empty
	if key == "" {
		return "", "", false
	}

	return key, value, true
}

// isNumericLike checks if value looks like a number (int/float)
func isNumericLike(value string) bool {
	// Pure integer
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return true
	}

	// Float
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return true
	}

	return false
}

// shouldQuoteValue determines if a value needs quoting based on content
func shouldQuoteValue(value string) bool {
	// Empty values should be quoted
	if value == "" {
		return true
	}

	// Quote if contains spaces
	if strings.ContainsAny(value, " \t") {
		return true
	}

	// Quote if looks purely numeric (to prevent type coercion)
	if isNumericLike(value) {
		return true
	}

	// Quote if contains special YAML characters
	specialChars := "{}[],:&*#?|-<>=!%@\\"
	if strings.ContainsAny(value, specialChars) {
		return true
	}

	// Quote if starts with quote character
	if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
		return true
	}

	// Quote YAML boolean-like values to preserve as strings
	lowerValue := strings.ToLower(value)
	yamlBools := []string{"true", "false", "yes", "no", "on", "off", "y", "n"}
	for _, b := range yamlBools {
		if lowerValue == b {
			return true
		}
	}

	// Otherwise, safe to leave unquoted
	return false
}

// getKeyOrder returns the sort order for docker-compose keys
// Lower numbers come first
//
// Ordering Philosophy:
// This implementation follows the Docker Compose specification's logical groupings
// while optimizing for developer experience and visual scanning. The spec provides
// implicit groupings but no strict ordering requirements.
//
// Top-level: Core metadata → Infrastructure → Services (longest section last)
// Service-level: What you define → How it runs → How it connects → Resources
//
// We use "priority bands" (gaps of 10) to allow easy insertion of new properties
// without renumbering. Properties within bands are sorted alphabetically.
func getKeyOrder(key string, isTopLevel bool) int {
	// Top-level keys order
	// Spec grouping: Core (version, name) → Infrastructure (networks, volumes, etc.) → Services
	// Services placed last as it's typically the longest and most complex section
	topLevelOrder := map[string]int{
		"version":  1,
		"name":     2,
		"include":  3,
		"networks": 10,
		"volumes":  20,
		"configs":  30,
		"secrets":  40,
		"models":   50,
		"services": 1000, // Services last at top level
	}

	// Service-level keys order
	// Organized by developer workflow: Identity → Execution → Configuration → Connectivity → Dependencies → Lifecycle → Resources
	// This differs slightly from the spec's alphabetical listing but follows logical usage patterns
	serviceLevelOrder := map[string]int{
		// Container Identity & Source (1-10)
		// What container to run and how to identify it
		"image":          1,
		"build":          2,
		"container_name": 3,
		"hostname":       4,
		"domainname":     5,
		"platform":       6,

		// Runtime Execution (10-20)
		// How the container runs and initializes
		"command":    10,
		"entrypoint": 11,
		"init":       12,

		// Environment Configuration (20-30)
		// Environment variables and configuration
		"environment": 20,
		"env_file":    21,

		// Port Exposure (30-40)
		// External port mappings and exposure
		"ports":  30,
		"expose": 31,

		// Storage & Data (40-50)
		// Volume mounts, devices, and temporary filesystems
		"volumes":      40,
		"volumes_from": 41,
		"devices":      42,
		"tmpfs":        43,

		// Network Configuration (50-60)
		// Network connections, links, and addressing
		"networks":       50,
		"network_mode":   51,
		"links":          52,
		"external_links": 53,
		"mac_address":    54,

		// Dependencies & Configuration Data (60-70)
		// Service dependencies and configuration/secret access
		"depends_on": 60,
		"secrets":    65,
		"configs":    66,

		// Lifecycle Policies (70-90)
		// Restart behavior, deployment, and health monitoring
		"restart":            70,
		"pull_policy":        71,
		"pull_refresh_after": 72,
		"deploy":             80,
		"develop":            81,
		"healthcheck":        90,

		// Metadata & Observability (100-120)
		// Labels, annotations, and logging
		"labels":      100,
		"label_file":  101,
		"annotations": 102,
		"attach":      103,
		"logging":     110,

		// DNS & Networking Advanced (120-130)
		// DNS configuration and host mappings
		"extra_hosts": 120,
		"dns":         121,
		"dns_opt":     122,
		"dns_search":  123,

		// Security & Capabilities (130-150)
		// Security options, capabilities, and access control
		"security_opt":  130,
		"cap_add":       131,
		"cap_drop":      132,
		"privileged":    133,
		"userns_mode":   134,
		"cgroup":        135,
		"cgroup_parent": 136,
		"ipc":           137,
		"pid":           138,
		"uts":           139,
		"isolation":     140,
		"read_only":     141,

		// User Context (150-160)
		// User, working directory, and group settings
		"user":        150,
		"working_dir": 151,
		"group_add":   152,

		// Interactive Mode (160-170)
		// TTY and STDIN configuration
		"stdin_open": 160,
		"tty":        161,

		// Advanced Features (170-180)
		// Runtime, scaling, inheritance, and profiles
		"runtime":  170,
		"scale":    171,
		"extends":  172,
		"profiles": 173,

		// CPU Resources (180-190)
		// CPU allocation and limits
		"cpus":           180,
		"cpu_count":      181,
		"cpu_percent":    182,
		"cpu_shares":     183,
		"cpu_period":     184,
		"cpu_quota":      185,
		"cpu_rt_runtime": 186,
		"cpu_rt_period":  187,
		"cpuset":         188,

		// Memory Resources (190-200)
		// Memory limits and swapping
		"mem_limit":       190,
		"mem_reservation": 191,
		"mem_swappiness":  192,
		"memswap_limit":   193,

		// Process & I/O Resources (194-200)
		// Process limits, block I/O, and OOM behavior
		"pids_limit":       194,
		"blkio_config":     195,
		"oom_kill_disable": 196,
		"oom_score_adj":    197,

		// Shared Memory (200-210)
		"shm_size": 200,

		// Stop Behavior (210-220)
		// Container stop configuration
		"stop_signal":       210,
		"stop_grace_period": 211,

		// System Limits (220-230)
		// ulimits, sysctls, and storage
		"ulimits":     220,
		"sysctls":     221,
		"storage_opt": 222,

		// Advanced Security & Device Control (223-230)
		"device_cgroup_rules": 223,
		"credential_spec":     224,

		// GPU & AI/ML Support (225-240)
		// GPU access and AI model configuration
		"gpus":            225,
		"models":          230,
		"provider":        231,
		"use_api_socket":  232,

		// Lifecycle Hooks (240-250)
		// Post-start and pre-stop hooks
		"post_start": 240,
		"pre_stop":   241,
	}

	// Build configuration keys order
	buildOrder := map[string]int{
		"context":    1,
		"dockerfile": 2,
		"args":       3,
		"target":     4,
		"cache_from": 5,
		"labels":     6,
	}

	// Deploy configuration keys order
	deployOrder := map[string]int{
		"mode":            1,
		"replicas":        2,
		"placement":       3,
		"update_config":   4,
		"rollback_config": 5,
		"resources":       6,
		"restart_policy":  7,
		"labels":          8,
		"endpoint_mode":   9,
	}

	// Network configuration keys order
	networkOrder := map[string]int{
		"driver":      1,
		"driver_opts": 2,
		"enable_ipv4": 3,
		"enable_ipv6": 4,
		"ipam":        5,
		"external":    10,
		"internal":    11,
		"attachable":  12,
		"name":        20,
		"labels":      30,
	}

	// Volume configuration keys order
	volumeOrder := map[string]int{
		"driver":      1,
		"driver_opts": 2,
		"external":    10,
		"name":        20,
		"labels":      30,
	}

	// Secrets configuration keys order (top-level)
	secretsOrder := map[string]int{
		"file":        1,
		"environment": 2,
		"external":    10,
		"name":        20,
	}

	// Configs configuration keys order (top-level)
	configsOrder := map[string]int{
		"file":        1,
		"environment": 2,
		"content":     3,
		"external":    10,
		"name":        20,
	}

	// Check which order map to use based on common patterns
	// Only check top-level order when actually at top level
	if isTopLevel {
		if order, ok := topLevelOrder[key]; ok {
			return order
		}
	} else {
		// Check service-level order first for non-top-level keys
		if order, ok := serviceLevelOrder[key]; ok {
			return order
		}
	}
	if order, ok := buildOrder[key]; ok {
		return order
	}
	if order, ok := deployOrder[key]; ok {
		return order
	}
	if order, ok := networkOrder[key]; ok {
		return order
	}
	if order, ok := volumeOrder[key]; ok {
		return order
	}
	if order, ok := secretsOrder[key]; ok {
		return order
	}
	if order, ok := configsOrder[key]; ok {
		return order
	}

	// Default order for unknown keys (alphabetical sorting will apply)
	return 1000
}
