// Package plugingen generates a GitHub Copilot CLI plugin marketplace
// from the ATV Starter Kit's scaffold templates.
//
// The generator is the single source of truth for the plugins/ tree,
// marketplace.json, .github/plugin/marketplace.json, and
// .claude-plugin/marketplace.json.
// Templates under pkg/scaffold/templates/ remain authoritative for both
// the atv init scaffold path and the marketplace/source-install paths;
// this package projects them into the plugin formats Copilot expects.
//
// All output is deterministic: lists are sorted, paths are slash-
// normalized, line endings are LF. Re-running the generator on a
// clean tree must produce a byte-identical result so the CI drift
// check (`go run ./cmd/plugingen -check`) is reliable.
package plugingen

// PluginManifest models the plugin.json manifest at the root of each
// plugin directory. See:
// https://docs.github.com/en/copilot/reference/cli-plugin-reference#pluginjson
//
// Only the fields ATV uses are modelled. Optional fields use omitempty
// so generated manifests stay terse.
type PluginManifest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Author      *Author  `json:"author,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
	Repository  string   `json:"repository,omitempty"`
	License     string   `json:"license,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// Component path fields. Skills and Agents are slices because the
	// CLI accepts string or string[]; we always emit slices for
	// consistency.
	Skills []string `json:"skills,omitempty"`
	Agents []string `json:"agents,omitempty"`
}

// Author is shared between PluginManifest and Marketplace.Owner.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Marketplace models the marketplace.json file at .github/plugin/.
// See:
// https://docs.github.com/en/copilot/reference/cli-plugin-reference#marketplacejson
type Marketplace struct {
	Name     string             `json:"name"`
	Owner    Author             `json:"owner"`
	Metadata MarketplaceMeta    `json:"metadata"`
	Plugins  []MarketplaceEntry `json:"plugins"`
}

// MarketplaceMeta is the top-level metadata object.
//
// PluginRoot tells the CLI where to resolve relative `source` paths
// from. We set it to "./plugins" so each entry's source can be just
// the plugin directory name.
type MarketplaceMeta struct {
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	PluginRoot  string `json:"pluginRoot,omitempty"`
}

// MarketplaceEntry is one plugin entry inside Marketplace.Plugins.
//
// Source is the path relative to PluginRoot (or repo root if
// PluginRoot is empty). For ATV, Source is always the plugin dir name.
type MarketplaceEntry struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Category    string   `json:"category,omitempty"`
}

// SourceInstallMarketplace models the marketplace.json file consumed by
// VS Code's source-install AgentPlugin flow. It intentionally stays separate
// from the Copilot CLI marketplace model because the source-install catalog
// has a curated UX surface and CE-style metadata fields.
type SourceInstallMarketplace struct {
	Name     string                       `json:"name"`
	Owner    Author                       `json:"owner"`
	Metadata SourceInstallMarketplaceMeta `json:"metadata"`
	Plugins  []SourceInstallEntry         `json:"plugins"`
}

// SourceInstallMarketplaceMeta is the top-level metadata object for the
// curated source-install catalog.
type SourceInstallMarketplaceMeta struct {
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

// SourceInstallEntry is one plugin entry in the curated source-install catalog.
type SourceInstallEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Author      *Author  `json:"author,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Source      string   `json:"source"`
}
