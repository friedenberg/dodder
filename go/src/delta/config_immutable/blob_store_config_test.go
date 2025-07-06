package config_immutable

import (
	"bytes"
	"testing"
)

func TestDefaultLocalBlobStoreConfig(t *testing.T) {
	config := DefaultLocalBlobStoreConfig()

	if config.GetBlobStoreType() != BlobStoreTypeLocal {
		t.Errorf("Expected blob store type %s, got %s", BlobStoreTypeLocal, config.GetBlobStoreType())
	}

	if config.GetLockInternalFiles() != true {
		t.Errorf("Expected LockInternalFiles to be true")
	}
}

func TestBlobStoreConfigWriteRead(t *testing.T) {
	config := DefaultLocalBlobStoreConfig()

	var buf bytes.Buffer
	_, err := config.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Config was not written to buffer")
	}
}

func TestDefaultV2Config(t *testing.T) {
	config := DefaultV2()

	blobStores := config.GetBlobStores()
	if len(blobStores) != 1 {
		t.Errorf("Expected 1 blob store, got %d", len(blobStores))
	}

	defaultStore, exists := blobStores["default"]
	if !exists {
		t.Error("Expected 'default' blob store to exist")
	}

	if defaultStore.Type != BlobStoreTypeLocal {
		t.Errorf("Expected default blob store type %s, got %s", BlobStoreTypeLocal, defaultStore.Type)
	}

	if config.GetDefaultBlobStore() != "default" {
		t.Errorf("Expected default blob store name 'default', got %s", config.GetDefaultBlobStore())
	}
}
