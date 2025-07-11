package sku

import (
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type CommitOptions struct {
	StoreOptions
	ids.RepoId
	ids.Clock
	Proto
	DontAddMissingTags bool
	DontAddMissingType bool
}

type StreamIndexOptions struct {
	ForceLatest      bool
	AddToStreamIndex bool
}

type StoreOptions struct {
	StreamIndexOptions StreamIndexOptions
	AddToInventoryList bool
	ApplyProto         bool // TODO remove
	ApplyProtoType     bool // TODO remove
	MergeCheckedOut    bool
	RunHooks           bool
	UpdateTai          bool
	Validate           bool
}

func GetStoreOptionsRealizeWithProto() StoreOptions {
	return StoreOptions{
		ApplyProto: true,
		RunHooks:   true,
		UpdateTai:  true,
	}
}

func GetStoreOptionsRealizeSansProto() StoreOptions {
	return StoreOptions{
		RunHooks:  true,
		UpdateTai: true,
	}
}

func GetStoreOptionsReindex() StoreOptions {
	return StoreOptions{
		StreamIndexOptions: StreamIndexOptions{
			ForceLatest:      true,
			AddToStreamIndex: true,
		},
	}
}

func GetStoreOptionsImport() StoreOptions {
	return StoreOptions{
		AddToInventoryList: true,
		RunHooks:           true,
		UpdateTai:          true,
		Validate:           true,
	}
}

func GetStoreOptionsRemoteTransfer() StoreOptions {
	return StoreOptions{
		AddToInventoryList: true,
	}
}

func GetStoreOptionsUpdate() StoreOptions {
	return StoreOptions{
		AddToInventoryList: true,
		RunHooks:           true,
		UpdateTai:          true,
		Validate:           true,
	}
}

func GetStoreOptionsCreate() StoreOptions {
	return StoreOptions{
		AddToInventoryList: true,
		RunHooks:           true,
		ApplyProto:         true,
		UpdateTai:          true,
		Validate:           true,
	}
}
