package ids

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

const (
	// TODO figure out a more ergonomic way of incrementing and labeling as
	// latest
	TagTypeTomlV0 = "!toml-tag-v0"
	TagTypeTomlV1 = "!toml-tag-v1"
	TagTypeLuaV1  = "!lua-tag-v1"
	TagTypeLuaV2  = "!lua-tag-v2"

	TypeTypeTomlV0 = "!toml-type-v0"
	TypeTypeTomlV1 = "!toml-type-v1"

	ConfigTypeTomlV0 = "!toml-config-v0"
	ConfigTypeTomlV1 = "!toml-config-v1"

	InventoryListTypeV0       = "!inventory_list-v0"
	InventoryListTypeV1       = "!inventory_list-v1"
	InventoryListTypeV2       = "!inventory_list-v2"
	InventoryListTypeVCurrent = InventoryListTypeV2

	RepoTypeXDGDotenvV0 = "!toml-repo-dotenv_xdg-v0"
	RepoTypeLocalPath   = "!toml-repo-local_path-v0"
	RepoTypeUri         = "!toml-repo-uri-v0"

	ImmutableConfigV1 = "!toml-config-immutable-v1"
	ImmutableConfigV2 = "!toml-config-immutable-v2"

	ZettelIdListTypeV0 = "!zettel_id_list-v0"

	WorkspaceConfigTypeTomlV0 = "!toml-workspace_config-v0"
)

type BuiltinType struct {
	Type
	genres.Genre
	Default bool
}

var (
	allSlice []BuiltinType
	allMap   map[Type]BuiltinType
	defaults map[genres.Genre]BuiltinType
)

func init() {
	allMap = make(map[Type]BuiltinType)
	defaults = make(map[genres.Genre]BuiltinType)

	registerBuiltinTypeString(TagTypeTomlV0, genres.Tag, false)
	registerBuiltinTypeString(TagTypeTomlV1, genres.Tag, true)
	registerBuiltinTypeString(TagTypeLuaV1, genres.Tag, false)
	registerBuiltinTypeString(TagTypeLuaV2, genres.Tag, false)

	registerBuiltinTypeString(TypeTypeTomlV0, genres.Type, false)
	registerBuiltinTypeString(TypeTypeTomlV1, genres.Type, true)

	registerBuiltinTypeString(ConfigTypeTomlV0, genres.Config, false)
	registerBuiltinTypeString(ConfigTypeTomlV1, genres.Config, true)

	registerBuiltinTypeString(InventoryListTypeV0, genres.InventoryList, false)
	registerBuiltinTypeString(InventoryListTypeV1, genres.InventoryList, false)
	// TODO StoreVersionV10
	registerBuiltinTypeString(InventoryListTypeV2, genres.InventoryList, true)

	registerBuiltinTypeString(RepoTypeUri, genres.Repo, true)
	registerBuiltinTypeString(RepoTypeXDGDotenvV0, genres.Repo, false)
	registerBuiltinTypeString(RepoTypeLocalPath, genres.Repo, false)

	registerBuiltinTypeString(ImmutableConfigV1, genres.None, false)
	registerBuiltinTypeString(ImmutableConfigV2, genres.None, false)

	registerBuiltinTypeString(ZettelIdListTypeV0, genres.None, false)

	registerBuiltinTypeString(WorkspaceConfigTypeTomlV0, genres.None, false)
}

// TODO switch to isDefault being a StoreVersion
func registerBuiltinTypeString(
	tipeString string,
	genre genres.Genre,
	isDefault bool,
) {
	registerBuiltinType(
		BuiltinType{
			Type:    MustType(tipeString),
			Genre:   genre,
			Default: isDefault,
		},
	)
}

func registerBuiltinType(bt BuiltinType) {
	if _, exists := allMap[bt.Type]; exists {
		panic(
			fmt.Sprintf("builtin type registered more than once: %s", bt.Type),
		)
	}

	if _, exists := defaults[bt.Genre]; exists && bt.Default {
		panic(
			fmt.Sprintf(
				"builtin default type registered more than once: %s",
				bt.Type,
			),
		)
	}

	allMap[bt.Type] = bt
	allSlice = append(allSlice, bt)

	if bt.Default {
		defaults[bt.Genre] = bt
	}
}

func IsBuiltin(tipe Type) bool {
	_, ok := allMap[tipe]
	return ok
}

func Get(t Type) (BuiltinType, bool) {
	bt, ok := allMap[t]
	return bt, ok
}

func GetOrPanic(idString string) BuiltinType {
	t := MustType(idString)
	bt, ok := Get(t)

	if !ok {
		panic(fmt.Sprintf("no builtin type found for %q", t))
	}

	return bt
}

func Default(genre genres.Genre) (Type, bool) {
	bt, ok := defaults[genre]
	return bt.Type, ok
}

func DefaultOrPanic(genre genres.Genre) Type {
	t, ok := Default(genre)

	if !ok {
		panic(fmt.Sprintf("default missing for genre %q", genre))
	}

	return t
}
