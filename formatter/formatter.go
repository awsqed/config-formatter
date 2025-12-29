package formatter

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// Format formats a docker-compose YAML file with consistent indentation and ordering
func Format(data []byte, indent int) ([]byte, error) {
	var root yaml.Node
	err := yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Apply formatting to the node tree
	formatNode(&root, true)

	// Marshal back to YAML with specified indentation
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(indent)

	err = encoder.Encode(&root)
	if err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}
	encoder.Close()

	// Post-process to fix empty lines (remove trailing spaces)
	result := cleanEmptyLines(buf.Bytes())

	return result, nil
}

// cleanEmptyLines removes trailing spaces from empty lines and removes leading empty lines
func cleanEmptyLines(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))

	// Remove trailing spaces from empty lines
	for i, line := range lines {
		// If line only contains spaces, make it truly empty
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 && len(line) > 0 {
			lines[i] = []byte{}
		}
	}

	// Remove leading empty lines
	start := 0
	for start < len(lines) && len(bytes.TrimSpace(lines[start])) == 0 {
		start++
	}
	if start > 0 {
		lines = lines[start:]
	}

	return bytes.Join(lines, []byte("\n"))
}

// formatNode recursively formats nodes in the YAML tree
func formatNode(node *yaml.Node, isRoot bool) {
	if node == nil {
		return
	}

	// Process mapping nodes (objects)
	if node.Kind == yaml.MappingNode {
		sortMappingNode(node, isRoot)
	}

	// Recursively format child nodes
	// Check if this is the root document node
	if isRoot && node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		formatNode(node.Content[0], true)
		return
	}

	for _, child := range node.Content {
		formatNode(child, false)
	}
}

// sortMappingNode sorts keys in a mapping node according to docker-compose conventions
func sortMappingNode(node *yaml.Node, isTopLevel bool) {
	if node.Kind != yaml.MappingNode || len(node.Content) == 0 {
		return
	}

	// Create pairs of key-value nodes
	type pair struct {
		key          *yaml.Node
		value        *yaml.Node
		order        int
		originalIdx  int
		hasComment   bool
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

// getKeyOrder returns the sort order for docker-compose keys
// Lower numbers come first
func getKeyOrder(key string, isTopLevel bool) int {
	// Top-level keys order - services comes last as it's usually the longest
	topLevelOrder := map[string]int{
		"version":  1,
		"name":     2,
		"networks": 10,
		"volumes":  20,
		"configs":  30,
		"secrets":  40,
		"services": 1000, // Services last at top level
	}

	// Service-level keys order
	serviceLevelOrder := map[string]int{
		"image":               1,
		"build":               2,
		"container_name":      3,
		"hostname":            4,
		"domainname":          5,
		"platform":            6,
		"command":             10,
		"entrypoint":          11,
		"init":                12,
		"environment":         20,
		"env_file":            21,
		"ports":               30,
		"expose":              31,
		"volumes":             40,
		"volumes_from":        41,
		"devices":             42,
		"tmpfs":               43,
		"networks":            50,
		"network_mode":        51,
		"links":               52,
		"external_links":      53,
		"mac_address":         54,
		"depends_on":          60,
		"secrets":             65,
		"configs":             66,
		"restart":             70,
		"pull_policy":         71,
		"deploy":              80,
		"develop":             81,
		"healthcheck":         90,
		"labels":              100,
		"label_file":          101,
		"annotations":         102,
		"attach":              103,
		"logging":             110,
		"extra_hosts":         120,
		"dns":                 121,
		"dns_opt":             122,
		"dns_search":          123,
		"security_opt":        130,
		"cap_add":             131,
		"cap_drop":            132,
		"privileged":          133,
		"userns_mode":         134,
		"cgroup":              135,
		"cgroup_parent":       136,
		"ipc":                 137,
		"pid":                 138,
		"uts":                 139,
		"isolation":           140,
		"read_only":           141,
		"user":                150,
		"working_dir":         151,
		"group_add":           152,
		"stdin_open":          160,
		"tty":                 161,
		"runtime":             170,
		"scale":               171,
		"extends":             172,
		"profiles":            173,
		"cpus":                180,
		"cpu_count":           181,
		"cpu_percent":         182,
		"cpu_shares":          183,
		"cpu_period":          184,
		"cpu_quota":           185,
		"cpu_rt_runtime":      186,
		"cpu_rt_period":       187,
		"cpuset":              188,
		"mem_limit":           190,
		"mem_reservation":     191,
		"mem_swappiness":      192,
		"memswap_limit":       193,
		"pids_limit":          194,
		"blkio_config":        195,
		"oom_kill_disable":    196,
		"oom_score_adj":       197,
		"shm_size":            200,
		"stop_signal":         210,
		"stop_grace_period":   211,
		"ulimits":             220,
		"sysctls":             221,
		"storage_opt":         222,
		"device_cgroup_rules": 223,
		"credential_spec":     224,
		"gpus":                225,
		"models":              230,
		"provider":            231,
		"use_api_socket":      232,
		"post_start":          240,
		"pre_stop":            241,
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
		"driver":       1,
		"driver_opts":  2,
		"enable_ipv4":  3,
		"enable_ipv6":  4,
		"ipam":         5,
		"external":     10,
		"internal":     11,
		"attachable":   12,
		"name":         20,
		"labels":       30,
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
	if order, ok := topLevelOrder[key]; ok {
		return order
	}
	if order, ok := serviceLevelOrder[key]; ok {
		return order
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
