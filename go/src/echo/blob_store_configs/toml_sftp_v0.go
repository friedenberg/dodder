package blob_store_configs

import (
	"flag"
)

type TomlSftpV0 struct {
	// TODO replace the below with a url scheme
	Host           string `toml:"host"`
	Port           int    `toml:"port,omitempty"`
	User           string `toml:"user"`
	Password       string `toml:"password,omitempty"`
	PrivateKeyPath string `toml:"private-key-path,omitempty"`
	RemotePath     string `toml:"remote-path"`
}

func (blobStoreConfig *TomlSftpV0) SetFlagSet(flagSet *flag.FlagSet) {
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
