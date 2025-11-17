package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/foxtrot/file_extensions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type DefaultsV1 struct {
	Type ids.Type  `toml:"type,omitempty"`
	Tags []ids.Tag `toml:"tags"`
}

var _ ConfigOverlay = V1{}

func (defaults DefaultsV1) GetDefaultType() ids.Type {
	return defaults.Type
}

func (defaults DefaultsV1) GetDefaultTags() collections_slice.Slice[ids.Tag] {
	return collections_slice.Slice[ids.Tag](defaults.Tags)
}

type DefaultsV1OmitEmpty struct {
	Type ids.Type  `toml:"type,omitempty"`
	Tags []ids.Tag `toml:"tags,omitempty"`
}

func (defaults DefaultsV1OmitEmpty) GetDefaultType() ids.Type {
	return defaults.Type
}

func (defaults DefaultsV1OmitEmpty) GetDefaultTags() collections_slice.Slice[ids.Tag] {
	return collections_slice.Slice[ids.Tag](defaults.Tags)
}

type V1 struct {
	Defaults       DefaultsV1             `toml:"defaults"`
	FileExtensions file_extensions.TOMLV1 `toml:"file-extensions"`
	PrintOptions   options_print.V1       `toml:"cli-output"`
	Tools          options_tools.Options  `toml:"tools"`
}

var _ ConfigOverlay = V1{}

func (blob *V1) Reset() {
	blob.FileExtensions.Reset()
	blob.Defaults.Type = ids.Type{}
	blob.Defaults.Tags = make([]ids.Tag, 0)
	blob.PrintOptions = options_print.V1{}
}

func (blob *V1) ResetWith(b *V1) {
	blob.FileExtensions.Reset()

	blob.Defaults.Type = b.Defaults.Type

	blob.Defaults.Tags = make([]ids.Tag, len(b.Defaults.Tags))
	copy(blob.Defaults.Tags, b.Defaults.Tags)

	blob.PrintOptions = b.PrintOptions
}

func (blob V1) GetDefaults() Defaults {
	return blob.Defaults
}

func (blob V1) GetFileExtensionsOverlay() file_extensions.Overlay {
	return blob.FileExtensions.GetFileExtensionsOverlay()
}

func (blob V1) GetPrintOptionsOverlay() options_print.Overlay {
	return blob.PrintOptions.GetPrintOptionsOverlay()
}

func (blob V1) GetToolOptions() options_tools.Options {
	return blob.Tools
}
