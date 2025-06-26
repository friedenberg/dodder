#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"
}

teardown() {
	rm_from_version "$version"
}

function format_simple { # @test
	run_zit_init_workspace
	run_zit checkout !md:t
	assert_success

	cat >md.type <<-EOM
		inline-akte = true
		[formatters.text]
		shell = [
		  "cat",
		]
	EOM

	# run cat .zit/Objekten/Akten/*/*
	# assert_output ''

	run_zit checkin -delete .t
	assert_success
	assert_output - <<-EOM
		[!md @21759bebd1a7937005f692b9394c0d2629361286b9fe837617e166c3ded687eb !toml-type-v1]
		          deleted [md.type]
	EOM

	run_zit format-object -mode both one/uno text
	assert_success
	assert_output - <<-EOM
		---
		# wow the first
		- tag-3
		- tag-4
		! md
		---

		last time
	EOM

	run_zit checkout one/uno
	assert_success
	cat >one/uno.zettel <<-EOM
		---
		# wow the second
		- tag-3
		- tag-4
		! md
		---

		last time but new
	EOM

	run_zit format-object -mode both one/uno.zettel text
	assert_success
	assert_output - <<-EOM
		---
		# wow the second
		- tag-3
		- tag-4
		! md
		---

		last time but new
	EOM
}

function show_simple_one_zettel_binary { # @test
	run_zit init-workspace
	assert_success

	echo "binary file" >file.bin
	run_zit add -delete file.bin
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [file.bin]
		[!bin @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
		[two/uno @b20c8fea8cb3e467783c5cdadf0707124cac5db72f9a6c3abba79fa0a42df627 !bin "file"]
	EOM

	run_zit checkout !bin:t
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [bin.type @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
	EOM

	cat >bin.type <<-EOM
		---
		! toml-type-v1
		---

		binary = true
	EOM

	run_zit checkin -delete bin.type
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [bin.type]
		[!bin @e07d72a74e0a01c23ddeb871751f6fcb43afec5fb81108c157537db96c6c1da0 !toml-type-v1]
	EOM

	run_zit format-object -mode both two/uno
	assert_success
	assert_output - <<-EOM
		---
		# file
		! b20c8fea8cb3e467783c5cdadf0707124cac5db72f9a6c3abba79fa0a42df627.bin
		---
	EOM
}
