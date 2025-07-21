package interfaces

type FileExtensionsGetter interface {
	GetFileExtensions() FileExtensions
}

type FileExtensions interface {
	GetFileExtensionForGenre(GenreGetter) string
	GetFileExtensionZettel() string
	GetFileExtensionOrganize() string
	GetFileExtensionType() string
	GetFileExtensionTag() string
	GetFileExtensionRepo() string
	GetFileExtensionConfig() string
}

type ObjectIOFactory interface {
	ObjectReaderFactory
	ObjectWriterFactory
}

type ObjectReaderFactory interface {
	ObjectReader(DigestGetter) (ReadCloserDigester, error)
}

type ObjectWriterFactory interface {
	ObjectWriter() (WriteCloserDigester, error)
}

type (
	FuncObjectReader func(DigestGetter) (ReadCloserDigester, error)
	FuncObjectWriter func() (WriteCloserDigester, error)
)

type bespokeObjectReadWriterFactory struct {
	ObjectReaderFactory
	ObjectWriterFactory
}

func MakeBespokeObjectReadWriterFactory(
	r ObjectReaderFactory,
	w ObjectWriterFactory,
) ObjectIOFactory {
	return bespokeObjectReadWriterFactory{
		ObjectReaderFactory: r,
		ObjectWriterFactory: w,
	}
}

type bespokeObjectReadFactory struct {
	FuncObjectReader
}

func MakeBespokeObjectReadFactory(
	r FuncObjectReader,
) ObjectReaderFactory {
	return bespokeObjectReadFactory{
		FuncObjectReader: r,
	}
}

func (b bespokeObjectReadFactory) ObjectReader(
	sh DigestGetter,
) (ReadCloserDigester, error) {
	return b.FuncObjectReader(sh)
}

type bespokeObjectWriterFactory struct {
	FuncObjectWriter
}

func MakeBespokeObjectWriteFactory(
	r FuncObjectWriter,
) ObjectWriterFactory {
	return bespokeObjectWriterFactory{
		FuncObjectWriter: r,
	}
}

func (b bespokeObjectWriterFactory) ObjectWriter() (WriteCloserDigester, error) {
	return b.FuncObjectWriter()
}
