package blob_store_config

import "code.linenisgreat.com/dodder/go/src/delta/compression_type"

type Current = BlobStoreTomlV1

func Default() Current {
	return Current{
		CompressionType:   compression_type.CompressionTypeDefault,
		LockInternalFiles: true,
	}
}
