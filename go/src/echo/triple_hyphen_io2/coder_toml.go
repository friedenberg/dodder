package triple_hyphen_io2

// TODO figure out how to implement this using generics
// type CoderToml[
// 	BLOB any,
// 	BLOB_PTR interfaces.Ptr[BLOB],
// ] struct{}

// func (CoderToml[BLOB, BLOB_PTR]) DecodeFrom(
// 	blob *BLOB_PTR,
// 	bufferedReader *bufio.Reader,
// ) (n int64, err error) {
// 	config := &genesis_config.TomlV1Private{}
// 	tomlDecoder := toml.NewDecoder(bufferedReader)

// 	if err = tomlDecoder.Decode(config); err != nil {
// 		if err == io.EOF {
// 			err = nil
// 		} else {
// 			err = errors.Wrap(err)
// 			return
// 		}
// 	}

// 	*blob = config

// 	return
// }

// func (CoderToml[BLOB, BLOB_PTR]) EncodeTo(
// 	blob *genesis_config.Private,
// 	bufferedWriter *bufio.Writer,
// ) (n int64, err error) {
// 	tomlEncoder := toml.NewEncoder(bufferedWriter)

// 	if err = tomlEncoder.Encode(*blob); err != nil {
// 		if err == io.EOF {
// 			err = nil
// 		} else {
// 			err = errors.Wrap(err)
// 			return
// 		}
// 	}

// 	return
// }
