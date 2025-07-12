package ids

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

const (
	// TODO figure out a more ergonomic way of incrementing and labeling as
	// latest

	// keep sorted
	TypeInventoryListV0       = "!inventory_list-v0"
	TypeInventoryListV1       = "!inventory_list-v1"
	TypeInventoryListV2       = "!inventory_list-v2"
	TypeInventoryListVCurrent = TypeInventoryListV2
	TypeLuaTagV1              = "!lua-tag-v1"
	TypeLuaTagV2              = "!lua-tag-v2"
	TypeTomlBlobStoreConfigV0 = "!toml-blob_store_config-v0"
	TypeTomlConfigImmutableV1 = "!toml-config-immutable-v1"
	TypeTomlConfigImmutableV2 = "!toml-config-immutable-v2"
	TypeTomlConfigV0          = "!toml-config-v0"
	TypeTomlConfigV1          = "!toml-config-v1"
	TypeTomlRepoDotenvXdgV0   = "!toml-repo-dotenv_xdg-v0"
	TypeTomlRepoLocalPath     = "!toml-repo-local_path-v0"
	TypeTomlRepoUri           = "!toml-repo-uri-v0"
	TypeTomlTagV0             = "!toml-tag-v0"
	TypeTomlTagV1             = "!toml-tag-v1"
	TypeTomlTypeV0            = "!toml-type-v0"
	TypeTomlTypeV1            = "!toml-type-v1"
	TypeTomlWorkspaceConfigV0 = "!toml-workspace_config-v0"
	TypeZettelIdListV0        = "!zettel_id_list-v0"
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

	// keep sorted
	registerBuiltinTypeString(TypeInventoryListV0, genres.InventoryList, false)
	registerBuiltinTypeString(TypeInventoryListV1, genres.InventoryList, false)
	registerBuiltinTypeString(TypeInventoryListV2, genres.InventoryList, true)
	registerBuiltinTypeString(TypeLuaTagV1, genres.Tag, false)
	registerBuiltinTypeString(TypeLuaTagV2, genres.Tag, false)
	registerBuiltinTypeString(TypeTomlBlobStoreConfigV0, genres.None, false)
	registerBuiltinTypeString(TypeTomlConfigImmutableV1, genres.None, false)
	registerBuiltinTypeString(TypeTomlConfigImmutableV2, genres.None, false)
	registerBuiltinTypeString(TypeTomlConfigV0, genres.Config, false)
	registerBuiltinTypeString(TypeTomlConfigV1, genres.Config, true)
	registerBuiltinTypeString(TypeTomlRepoDotenvXdgV0, genres.Repo, false)
	registerBuiltinTypeString(TypeTomlRepoLocalPath, genres.Repo, false)
	registerBuiltinTypeString(TypeTomlRepoUri, genres.Repo, true)
	registerBuiltinTypeString(TypeTomlTagV0, genres.Tag, false)
	registerBuiltinTypeString(TypeTomlTagV1, genres.Tag, true)
	registerBuiltinTypeString(TypeTomlTypeV0, genres.Type, false)
	registerBuiltinTypeString(TypeTomlTypeV1, genres.Type, true)
	registerBuiltinTypeString(TypeTomlWorkspaceConfigV0, genres.None, false)
	registerBuiltinTypeString(TypeZettelIdListV0, genres.None, false)
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
