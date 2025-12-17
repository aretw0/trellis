package dto

// NodeMetadata represents the header/metadata of a Trellis Node.
// It uses "mapstructure" tags to match standard Frontmatter/YAML keys (to, from).
type NodeMetadata struct {
	ID          string             `json:"id" mapstructure:"id"`
	Type        string             `json:"type" mapstructure:"type"`
	Transitions []LoaderTransition `json:"transitions" mapstructure:"transitions"`
}

type LoaderTransition struct {
	From      string `json:"from" mapstructure:"from"`
	FromFull  string `json:"from_node_id" mapstructure:"from_node_id"`
	To        string `json:"to" mapstructure:"to"`
	ToFull    string `json:"to_node_id" mapstructure:"to_node_id"`
	Condition string `json:"condition" mapstructure:"condition"`
}
