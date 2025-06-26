package env_dir

// func Test1(t *testing.T) {
// 	var err error
// 	age := makeAge(t)

// 	text := `test string`
// 	in := strings.NewReader(text)
// 	out := &strings.Builder{}

// 	var w *writer

// 	o := WriteOptions{
// 		Config: MakeConfig(
// 			age,
// 			immutable_config.CompressionTypeDefault,
// 			false,
// 		),
// 		Writer: out,
// 	}

// 	if w, err = NewWriter(o); err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	if _, err = io.Copy(w, in); err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	w.Close()

// 	in = strings.NewReader(out.String())
// 	out = &strings.Builder{}

// 	var r *reader

// 	ro := ReadOptions{
// 		Config: MakeConfig(
// 			age,
// 			immutable_config.CompressionTypeDefault,
// 			false,
// 		),
// 		Reader: in,
// 	}

// 	if r, err = NewReader(ro); err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	if _, err = io.Copy(out, r); err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	if text != out.String() {
// 		t.Fatalf("expected '%s', but got '%s'", text, out.String())
// 	}
// }
