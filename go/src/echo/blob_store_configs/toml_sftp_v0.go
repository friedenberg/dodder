package blob_store_configs

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type TomlSftpV0 struct {
	// TODO replace the below with a url scheme
	Host           string `toml:"host"`
	Port           int    `toml:"port,omitempty"`
	User           string `toml:"user"`
	Password       string `toml:"password,omitempty"`
	PrivateKeyPath string `toml:"private-key-path,omitempty"`
	RemotePath     string `toml:"remote-path"`

	// TODO modify blob store config to read this after blob store
	// initialization
	AgeEncryption     age.Age                          `toml:"age-encryption,omitempty"`
	CompressionType   compression_type.CompressionType `toml:"compression-type"`
	LockInternalFiles bool                             `toml:"lock-internal-files"`
}

func (blobStoreConfig *TomlSftpV0) SetFlagSet(flagSet *flag.FlagSet) {
	blobStoreConfig.CompressionType.SetFlagSet(flagSet)

	flagSet.StringVar(
		&blobStoreConfig.Host,
		"sftp-host",
		blobStoreConfig.Host,
		"SFTP server hostname",
	)

	flagSet.IntVar(
		&blobStoreConfig.Port,
		"sftp-port",
		22,
		"SFTP server port",
	)

	flagSet.StringVar(
		&blobStoreConfig.User,
		"sftp-user",
		blobStoreConfig.User,
		"SFTP username",
	)

	flagSet.StringVar(
		&blobStoreConfig.Password,
		"sftp-password",
		blobStoreConfig.Password,
		"SFTP password",
	)

	flagSet.StringVar(
		&blobStoreConfig.PrivateKeyPath,
		"sftp-private-key-path",
		blobStoreConfig.PrivateKeyPath,
		"Path to SSH private key",
	)

	flagSet.StringVar(
		&blobStoreConfig.RemotePath,
		"sftp-remote-path",
		blobStoreConfig.RemotePath,
		"Remote path for blob storage",
	)

	flagSet.BoolVar(
		&blobStoreConfig.LockInternalFiles,
		"lock-internal-files",
		blobStoreConfig.LockInternalFiles,
		"",
	)

	flagSet.Var(
		&blobStoreConfig.AgeEncryption,
		"age-identity",
		"add an age identity",
	)
}

func (blobStoreConfig *TomlSftpV0) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return blobStoreConfig
}

func (blobStoreConfig *TomlSftpV0) GetBlobCompression() interfaces.BlobCompression {
	return &blobStoreConfig.CompressionType
}

func (blobStoreConfig *TomlSftpV0) GetBlobEncryption() interfaces.BlobEncryption {
	return &blobStoreConfig.AgeEncryption
}

func (blobStoreConfig *TomlSftpV0) GetLockInternalFiles() bool {
	return blobStoreConfig.LockInternalFiles
}

func (blobStoreConfig *TomlSftpV0) GetHost() string {
	return blobStoreConfig.Host
}

func (blobStoreConfig *TomlSftpV0) GetPort() int {
	if blobStoreConfig.Port == 0 {
		return 22
	}
	return blobStoreConfig.Port
}

func (blobStoreConfig *TomlSftpV0) GetUser() string {
	return blobStoreConfig.User
}

func (blobStoreConfig *TomlSftpV0) GetPassword() string {
	return blobStoreConfig.Password
}

func (blobStoreConfig *TomlSftpV0) GetPrivateKeyPath() string {
	return blobStoreConfig.PrivateKeyPath
}

func (blobStoreConfig *TomlSftpV0) GetRemotePath() string {
	return blobStoreConfig.RemotePath
}
