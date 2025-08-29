package merkle

// TODO move to ids' builtin types
// and then add registration

const (
	// keep sorted
	HRPObjectBlobDigestSha256V0 = "sha256"
	HRPObjectBlobDigestSha256V1 = "dodder-object-blob-digest-sha256-v1"
	HRPObjectDigestSha256V1     = "dodder-object-digest-sha256-v1"
	HRPObjectMotherSigV1        = "dodder-object-mother-sig-v1"
	HRPObjectSigV0              = "dodder-repo-sig-v1"
	HRPObjectSigV1              = "dodder-object-sig-v1"
	HRPRepoPrivateKeyV1         = "dodder-repo-private_key-v1"
	HRPRepoPubKeyV1             = "dodder-repo-public_key-v1"
	HRPRequestAuthChallengeV1   = "dodder-request_auth-challenge-v1"
	HRPRequestAuthResponseV1    = "dodder-request_auth-response-v1"
)

var hrpValid = []string{
	// keep sorted
	HRPObjectBlobDigestSha256V0,
	HRPObjectBlobDigestSha256V1,
	HRPObjectDigestSha256V1,
	HRPObjectMotherSigV1,
	HRPObjectSigV0,
	HRPObjectSigV1,
	HRPRepoPrivateKeyV1,
	HRPRepoPubKeyV1,
	HRPRequestAuthChallengeV1,
	HRPRequestAuthResponseV1,
}
