#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"
	# TODO prevent checkouts if workspace is not initialized
	run_zit_init_workspace

	cat >txt.type <<-EOM
		---
		! toml-type-v1
		---

		binary = false
	EOM

	cat >bin.type <<-EOM
		---
		! toml-type-v1
		---

		binary = true
	EOM

	run_zit checkin -delete bin.type txt.type
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [bin.type]
		          deleted [txt.type]
		[!bin @e07d72a74e0a01c23ddeb871751f6fcb43afec5fb81108c157537db96c6c1da0 !toml-type-v1]
		[!txt @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
	EOM
}

teardown() {
	rm_from_version "$version"
}

function checkout_simple_all { # @test
	run_zit checkout :z,t,e
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [txt.type @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
		      checked out [bin.type @e07d72a74e0a01c23ddeb871751f6fcb43afec5fb81108c157537db96c6c1da0 !toml-type-v1]
		      checked out [md.type @$(get_type_blob_sha) !toml-type-v1]
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		      checked out [tag-1.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag-2.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag-3.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag-4.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

function checkout_simple_zettel { # @test
	run_zit checkout :
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function checkout_non_binary_simple_zettel { # @test
	echo "text file" >file.txt
	run_zit add -delete file.txt
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [file.txt]
		[two/uno @f4796bff42910365a3c637e6e6dfc7cf78dba30387003e842e6043d3ec4923e3 !txt "file"]
	EOM

	run_zit show -format text !txt:z
	assert_success
	assert_output - <<-EOM
		---
		# file
		! txt
		---

		text file
	EOM
}

function checkout_binary_simple_zettel { # @test
	echo "binary file" >file.bin
	run_zit add -delete file.bin
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [file.bin]
		[two/uno @b20c8fea8cb3e467783c5cdadf0707124cac5db72f9a6c3abba79fa0a42df627 !bin "file"]
	EOM

	run_zit checkout !bin:z
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [two/uno.zettel @b20c8fea8cb3e467783c5cdadf0707124cac5db72f9a6c3abba79fa0a42df627 !bin "file"]
	EOM

	run cat two/uno.zettel
	assert_success
	assert_output - <<-EOM
		---
		# file
		! b20c8fea8cb3e467783c5cdadf0707124cac5db72f9a6c3abba79fa0a42df627.bin
		---
	EOM
}

function checkout_simple_zettel_blob_only { # @test
	run_zit clean .
	assert_success
	# TODO fail checkouts if working directly has incompatible checkout
	run_zit checkout -mode blob :z
	assert_success
	assert_output_unsorted - <<-EOM
		                   one/dos.md]
		                   one/uno.md]
		      checked out [one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4
		      checked out [one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4
	EOM
}

function checkout_zettel_several { # @test
	run_zit checkout one/uno one/dos
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function checkout_simple_type { # @test
	run_zit checkout :t
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [bin.type @e07d72a74e0a01c23ddeb871751f6fcb43afec5fb81108c157537db96c6c1da0 !toml-type-v1]
		      checked out [md.type @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		      checked out [txt.type @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
	EOM
}

function checkout_zettel_blob_then_object { # @test
	run_zit checkout -mode blob one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4
		                   one/uno.md]
	EOM

	run_zit checkout one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run ls one/
	assert_output_unsorted - <<-EOM
		uno.zettel
	EOM

	run_zit checkout -force one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run ls one/
	assert_output - <<-EOM
		uno.zettel
	EOM
}

function mode_both { # @test
	run_zit new -edit=false - <<-EOM
		---
		! bin
		---

		not really pdf content but that's ok
	EOM
	assert_success
	assert_output - <<-EOM
		[two/uno @22cbac1a49e4d3a94e97f18394400f8a0c99639293d8e2dc59eb14cbe8406e4e !bin]
	EOM

	run_zit checkout -mode both two/uno
	assert_success
	assert_output - <<-EOM
		      checked out [two/uno.zettel @22cbac1a49e4d3a94e97f18394400f8a0c99639293d8e2dc59eb14cbe8406e4e !bin
		                   two/uno.bin]
	EOM

	run ls two/
	assert_output_unsorted - <<-EOM
		uno.bin
		uno.zettel
	EOM
}

# bats test_tags=user_story:builtin_types
function checkout_builtin_type { # @test
	run_zit checkout !toml-type-v1:t
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [md.type @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		      checked out [txt.type @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
		      checked out [bin.type @e07d72a74e0a01c23ddeb871751f6fcb43afec5fb81108c157537db96c6c1da0 !toml-type-v1]
	EOM
}
