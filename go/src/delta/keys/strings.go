package keys

import (
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
)

const (
	ShaKeySelfMetadataWithouTai        = "SelbstMetadateiMutterSansTai"
	ShaKeySelfMetadataObjectIdParent   = "SelbstMetadateiKennungMutter"
	ShaKeyParentMetadataObjectIdParent = "MutterMetadateiKennungMutter"
	ShaKeySelf                         = "MetadateiTai"
	ShaKeyParent                       = "MutterMetadateiMutterTai"
)

var (
	KeySigil = catgut.MakeFromString("Sigil")
	KeyTai   = catgut.MakeFromString("Tai")

	KeySha = catgut.MakeFromString("zzSha")

	KeyBlob        = catgut.MakeFromString("Blob")
	KeyDescription = catgut.MakeFromString("Description")
	KeyTag         = catgut.MakeFromString("Tag")
	KeyGenre       = catgut.MakeFromString("Genre")
	KeyObjectId    = catgut.MakeFromString("ObjectId")
	KeyComment     = catgut.MakeFromString("Comment")
	KeyType        = catgut.MakeFromString("Type")

	// KeyCacheDormant     = catgut.MakeFromString("Cache-Dormant")
	// KeyCacheTagImplicit = catgut.MakeFromString("Cache-Tag-Implicit")
	// KeyCacheTagExpanded = catgut.MakeFromString("Cache-Tag-Expanded")
)
