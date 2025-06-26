package repo_signing

type Data []byte

func (data *Data) SetBytes(bs []byte) {
	data.Reset()
	*data = append(*data, bs...)
}

func (data *Data) GetBytes() []byte {
	return []byte(*data)
}

func (data *Data) IsEmpty() bool {
	return len(*data) == 0
}

func (data *Data) UnmarshalBinary(bytes []byte) (err error) {
	*data = bytes
	return
}

func (data *Data) MarshalBinary() (bytes []byte, err error) {
	bytes = []byte(*data)
	return
}

func (data *Data) Reset() {
	*data = (*data)[:0]
}

func (data *Data) ResetWith(src Data) {
	data.SetBytes(src.GetBytes())
}
