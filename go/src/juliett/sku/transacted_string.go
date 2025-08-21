package sku

import "fmt"

func (transacted *Transacted) String() string {
	return fmt.Sprintf(
		"%s %s %s",
		&transacted.ObjectId,
		transacted.GetObjectFingerPrint(),
		transacted.GetBlobId(),
	)
}

func (transacted *Transacted) StringObjectIdDescription() string {
	return fmt.Sprintf(
		"[%s %q]",
		&transacted.ObjectId,
		transacted.Metadata.Description,
	)
}

func (transacted *Transacted) StringObjectIdTai() string {
	return fmt.Sprintf(
		"%s@%s",
		&transacted.ObjectId,
		transacted.GetTai().StringDefaultFormat(),
	)
}

func (transacted *Transacted) StringObjectIdTaiBlob() string {
	return fmt.Sprintf(
		"%s@%s@%s",
		&transacted.ObjectId,
		transacted.GetTai().StringDefaultFormat(),
		transacted.GetBlobId(),
	)
}

func (transacted *Transacted) StringObjectIdSha() string {
	return fmt.Sprintf(
		"%s@%s",
		&transacted.ObjectId,
		transacted.GetMetadata().GetDigest(),
	)
}

func (transacted *Transacted) StringObjectIdParent() string {
	return fmt.Sprintf(
		"%s^@%s",
		&transacted.ObjectId,
		transacted.GetMetadata().GetMotherDigest(),
	)
}
