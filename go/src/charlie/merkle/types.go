package merkle

const (
	HRPRepoPubKeyV1             = "dodder-repo-public_key-v1"
	HRPRepoPrivateKeyV1         = "dodder-repo-private_key-v1"
	HRPRepoSigV1                = "dodder-repo-sig-v1"
	HRPRequestAuthChallengeV1   = "dodder-request_auth-challenge-v1"
	HRPRequestAuthResponseV1    = "dodder-request_auth-response-v1"
	HRPObjectBlobDigestSha256V0 = "sha256"
	HRPObjectBlobDigestSha256V1 = "dodder-object-blob-digest-sha256-v1"
	HRPObjectDigestSha256V1     = "dodder-object-digest-sha256-v1"
)

var hrpValid = []string{
	HRPRepoPubKeyV1,
	HRPRepoPrivateKeyV1,
	HRPRepoSigV1,
	HRPRequestAuthChallengeV1,
	HRPRequestAuthResponseV1,
	HRPObjectDigestSha256V1,
	HRPObjectBlobDigestSha256V0,
	HRPObjectBlobDigestSha256V1,
}
