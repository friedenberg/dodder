package sku

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

type UnsureMatchType byte

const (
	UnsureMatchTypeNone = UnsureMatchType(iota << 1)
	UnsureMatchTypeMetadataWithoutTaiHistory
	UnsureMatchTypeDescription

	UnsureMatchTypeAll = UnsureMatchType(^byte(0))
)

func (a UnsureMatchType) Contains(b UnsureMatchType) bool {
	return a&b == b
}

func (a UnsureMatchType) MakeMatchMap() UnsureMatchMaps {
	maps := UnsureMatchMaps{
		Lookup: make(map[UnsureMatchType]UnsureMatchMap),
	}

	if a.Contains(UnsureMatchTypeMetadataWithoutTaiHistory) {
		maps.Lookup[UnsureMatchTypeMetadataWithoutTaiHistory] = UnsureMatchMap{
			UnsureMatchType: UnsureMatchTypeMetadataWithoutTaiHistory,
			Lookup:          make(map[string]SkuTypeSetMutable),
		}
	}

	if a.Contains(UnsureMatchTypeDescription) {
		maps.Lookup[UnsureMatchTypeDescription] = UnsureMatchMap{
			UnsureMatchType: UnsureMatchTypeDescription,
			Lookup:          make(map[string]SkuTypeSetMutable),
		}
	}

	return maps
}

type UnsureMatchOptions struct {
	UnsureMatchType
}

func UnsureMatchOptionsDefault() UnsureMatchOptions {
	return UnsureMatchOptions{
		UnsureMatchType: UnsureMatchTypeMetadataWithoutTaiHistory | UnsureMatchTypeDescription,
	}
}

type IterMatching func(
	mt UnsureMatchType,
	sk *Transacted,
	existing SkuTypeSetMutable,
) error

type UnsureMatchMap struct {
	UnsureMatchType
	Lookup map[string]SkuTypeSetMutable
}

type UnsureMatchMaps struct {
	Lookup map[UnsureMatchType]UnsureMatchMap
}

func (umm UnsureMatchMaps) Len() int {
	l := 0

	for _, v := range umm.Lookup {
		l += len(v.Lookup)
	}

	return l
}

func MakeUnsureMatchMapsCollector(
	umm UnsureMatchMaps,
) interfaces.FuncIter[SkuType] {
	var l sync.Mutex

	return func(co SkuType) (err error) {
		e := co.GetSkuExternal().GetSku()

		l.Lock()
		defer l.Unlock()

		clone := co.Clone()

		for t, v := range umm.Lookup {
			var digestBytes []byte

			switch t {
			case UnsureMatchTypeMetadataWithoutTaiHistory:
				digestBytes = e.Metadata.Shas.SelfMetadataWithoutTai.GetBytes()

			case UnsureMatchTypeDescription:
				digestBytes = sha.FromStringContent(
					e.Metadata.Description.String(),
				).GetBytes()

			default:
				continue
			}

			existing, ok := v.Lookup[string(digestBytes)]

			if !ok {
				existing = MakeSkuTypeSetMutable()
			}

			if err = existing.Add(clone); err != nil {
				err = errors.Wrap(err)
				return
			}

			v.Lookup[string(digestBytes)] = existing
		}

		return
	}
}

func MakeUnsureMatchMapsMatcher(
	umm UnsureMatchMaps,
	f IterMatching,
) interfaces.FuncIter[*Transacted] {
	return func(sk *Transacted) (err error) {
		for t, v := range umm.Lookup {
			var digestBytes []byte

			switch t {
			case UnsureMatchTypeMetadataWithoutTaiHistory:
				digestBytes = sk.Metadata.Shas.SelfMetadataWithoutTai.GetBytes()

			case UnsureMatchTypeDescription:
				digestBytes = sha.FromStringContent(
					sk.Metadata.Description.String(),
				).GetBytes()

			default:
				continue
			}

			existing, ok := v.Lookup[string(digestBytes)]

			if !ok {
				continue
			}

			if err = f(
				t,
				sk,
				existing,
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		return
	}
}
