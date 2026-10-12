package assert

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// DecodeStrict decodes a node into v, rejecting keys v does not have.
func DecodeStrict(n *yaml.Node, v any) error {
	data, err := yaml.Marshal(n)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	return dec.Decode(v)
}
