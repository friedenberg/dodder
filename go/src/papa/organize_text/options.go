package organize_text

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/lima/box_format"
)

type Flags struct {
	Options

	once      *sync.Once
	ExtraTags collections_ptr.Flag[ids.Tag, *ids.Tag]
}

var _ interfaces.CommandComponentWriter = (*Flags)(nil)

type Options struct {
	wasMade bool

	Config interfaces.MutableConfigDryRun

	Metadata

	commentMatchers interfaces.Set[sku.Query]
	GroupingTags    ids.TagSlice
	ExtraTags       ids.TagSet
	Skus            sku.SkuTypeSet

	sku.ObjectFactory

	Abbr ids.Abbr

	UsePrefixJoints   bool
	UseRefiner        bool
	UseMetadataHeader bool
	Limit             int

	PrintOptions options_print.Options
	fmtBox       *box_format.BoxCheckedOut
}

func MakeFlags() Flags {
	return Flags{
		once: &sync.Once{},
		ExtraTags: collections_ptr.MakeFlagCommas[ids.Tag](
			collections_ptr.SetterPolicyAppend,
		),

		Options: Options{
			wasMade:      true,
			GroupingTags: ids.MakeTagSlice(),
			Skus:         sku.MakeSkuTypeSetMutable(),
			Metadata:     NewMetadata(ids.RepoId{}),
		},
	}
}

func MakeFlagsWithMetadata(m Metadata) Flags {
	if m.TagSet == nil {
		m.TagSet = ids.MakeTagSetFromSlice()
	}

	return Flags{
		once: &sync.Once{},
		ExtraTags: collections_ptr.MakeFlagCommas[ids.Tag](
			collections_ptr.SetterPolicyAppend,
		),

		Options: Options{
			Metadata:     m,
			wasMade:      true,
			GroupingTags: ids.MakeTagSlice(),
			Skus:         sku.MakeSkuTypeSetMutable(),
		},
	}
}

func (o *Flags) SetFlagDefinitions(f interfaces.CLIFlagDefinitions) {
	f.Var(&o.GroupingTags, "group-by", "tag prefixes to group zettels")

	f.Var(
		o.ExtraTags,
		"extras",
		"tags to always add to the organize text",
	)

	f.BoolVar(
		&o.UsePrefixJoints,
		"prefix-joints",
		true,
		"split tags around hyphens",
	)

	f.BoolVar(&o.UseRefiner, "refine", true, "refine the organize tree")

	f.BoolVar(
		&o.UseMetadataHeader,
		"metadata-header",
		true,
		"metadata header",
	)

	f.IntVar(
		&o.Limit,
		"limit",
		0,
		"limit the number of objects edited in organize",
	)
}

func (flags *Flags) GetOptionsWithMetadata(
	printOptions options_print.Options,
	boxFormat *box_format.BoxCheckedOut,
	abbr ids.Abbr,
	objectFactory sku.ObjectFactory,
	metadata Metadata,
) Options {
	flags.once.Do(
		func() {
			flags.Options.ExtraTags = ids.CloneTagSet(flags.ExtraTags)
		},
	)

	flags.fmtBox = boxFormat

	objectFactory.SetDefaultsIfNecessary()

	flags.ObjectFactory = objectFactory
	flags.PrintOptions = printOptions
	flags.Abbr = abbr
	flags.Metadata = metadata

	return flags.Options
}

func (o *Flags) GetOptions(
	printOptions options_print.Options,
	tagSet ids.TagSet,
	skuBoxFormat *box_format.BoxCheckedOut,
	abbr ids.Abbr, // TODO move Abbr as required arg
	objectFactory sku.ObjectFactory,
) Options {
	m := o.Metadata
	m.TagSet = tagSet

	if m.prototype == nil {
		panic("Metadata not initalized")
	}

	return o.GetOptionsWithMetadata(
		printOptions,
		skuBoxFormat,
		abbr,
		objectFactory,
		m,
	)
}

func (o Options) Make() (ot *Text, err error) {
	c := &constructor{
		Text: Text{
			Options: o,
		},
	}

	ot = &c.Text

	c.all = MakePrefixSet(0)
	c.Assignment = newAssignment(0)
	c.IsRoot = true

	if c.TagSet == nil {
		c.TagSet = ids.MakeTagSetFromSlice()
	}

	var objects Objects

	for sk := range c.Options.Skus.All() {
		objects.Add(&obj{sku: sk})
	}

	objects.Sort()

	for i, obj := range objects {
		if i != 0 && i == o.Limit {
			break
		}

		if err = c.all.AddSku(obj.sku); err != nil {
			err = errors.Wrap(err)
			return ot, err
		}
	}

	if err = c.preparePrefixSetsAndRootsAndExtras(); err != nil {
		err = errors.Wrap(err)
		return ot, err
	}

	if err = c.populate(); err != nil {
		err = errors.Wrap(err)
		return ot, err
	}

	c.Metadata.Type = c.Options.Type

	if err = ot.Refine(); err != nil {
		err = errors.Wrap(err)
		return ot, err
	}

	ot.SortChildren()

	return ot, err
}

func (o Options) refiner() *Refiner {
	return &Refiner{
		Enabled:         o.UseRefiner,
		UsePrefixJoints: o.UsePrefixJoints,
	}
}
