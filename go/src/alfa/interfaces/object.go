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
	ObjectReader(Digester) (ReadCloseDigester, error)
}

type ObjectWriterFactory interface {
	ObjectWriter() (WriteCloseDigester, error)
}

type (
	FuncObjectReader func(Digester) (ReadCloseDigester, error)
	FuncObjectWriter func() (WriteCloseDigester, error)
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
	sh Digester,
) (ReadCloseDigester, error) {
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

func (b bespokeObjectWriterFactory) ObjectWriter() (WriteCloseDigester, error) {
	return b.FuncObjectWriter()
}
