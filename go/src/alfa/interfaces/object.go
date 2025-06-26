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
	ObjectReader(ShaGetter) (ShaReadCloser, error)
}

type ObjectWriterFactory interface {
	ObjectWriter() (ShaWriteCloser, error)
}

type (
	FuncObjectReader func(ShaGetter) (ShaReadCloser, error)
	FuncObjectWriter func() (ShaWriteCloser, error)
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
	sh ShaGetter,
) (ShaReadCloser, error) {
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

func (b bespokeObjectWriterFactory) ObjectWriter() (ShaWriteCloser, error) {
	return b.FuncObjectWriter()
}
