package versions

import (
	"slices"

	"go.yaml.in/yaml/v4"
)

// field is one key of a mapping with its value.
type field struct {
	key   string
	line  int
	value *yaml.Node
}

// parse reads a YAML document and returns its value, or nil for an empty
// file. JSON is read the same way, which gives its values a line number.
func parse(content []byte) (*yaml.Node, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(content, &document); err != nil {
		return nil, err
	}
	if len(document.Content) == 0 {
		return nil, nil
	}
	return document.Content[0], nil
}

// fields lists the keys of a mapping. A node that is missing or is not a
// mapping has none.
func fields(node *yaml.Node) []field {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	var listed []field
	for index := 0; index+1 < len(node.Content); index += 2 {
		key := node.Content[index]
		listed = append(listed, field{key: key.Value, line: key.Line, value: node.Content[index+1]})
	}
	return listed
}

// child returns the value of a key of a mapping, or nil.
func child(node *yaml.Node, key string) *yaml.Node {
	listed := fields(node)
	index := slices.IndexFunc(listed, func(candidate field) bool { return candidate.key == key })
	if index < 0 {
		return nil
	}
	return listed[index].value
}

// isText tells whether a node is a single value, not a mapping or a list.
func isText(node *yaml.Node) bool {
	return node != nil && node.Kind == yaml.ScalarNode
}

// texts lists the single values a node holds: itself, or the items of a
// list.
func texts(node *yaml.Node) []*yaml.Node {
	if isText(node) {
		return []*yaml.Node{node}
	}
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	var listed []*yaml.Node
	for _, item := range node.Content {
		listed = append(listed, texts(item)...)
	}
	return listed
}

// mappings lists every mapping under a node, itself included, at any depth.
// An alias is not followed: the mapping it points at is listed where it is
// written.
func mappings(node *yaml.Node) []*yaml.Node {
	if node == nil {
		return nil
	}
	var listed []*yaml.Node
	if node.Kind == yaml.MappingNode {
		listed = append(listed, node)
	}
	for _, inner := range node.Content {
		listed = append(listed, mappings(inner)...)
	}
	return listed
}
