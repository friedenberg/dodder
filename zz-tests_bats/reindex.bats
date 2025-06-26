#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"
  run_zit_init_workspace
}

teardown() {
	rm_from_version "$version"
}

function reindex_simple { # @test
	run_zit reindex
	assert_success
	run_zit show +t,e,z,konfig
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[konfig @$(get_konfig_sha) !toml-config-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_zit show -format tags-path :e,z,t
	assert_success
	assert_output_unsorted - <<-EOM
		!md [Paths: [], All: []]
		one/dos [Paths: [TypeDirect:[tag-3] TypeDirect:[tag-4]], All: [tag-3:[TypeDirect:[tag-3]] tag-4:[TypeDirect:[tag-4]]]]
		one/uno [Paths: [TypeDirect:[tag-3] TypeDirect:[tag-4]], All: [tag-3:[TypeDirect:[tag-3]] tag-4:[TypeDirect:[tag-4]]]]
		tag [Paths: [TypeSelf:[tag]], All: [tag:[TypeSelf:[tag]]]]
		tag-1 [Paths: [TypeSelf:[tag-1]], All: [tag-1:[TypeSelf:[tag-1]]]]
		tag-2 [Paths: [TypeSelf:[tag-2]], All: [tag-2:[TypeSelf:[tag-2]]]]
		tag-3 [Paths: [TypeSelf:[tag-3]], All: [tag-3:[TypeSelf:[tag-3]]]]
		tag-4 [Paths: [TypeSelf:[tag-4]], All: [tag-4:[TypeSelf:[tag-4]]]]
	EOM
}

function reindex_simple_twice { # @test
	expected="$(mktemp)"
	cat - >"$expected" <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[konfig @$(get_konfig_sha) !toml-config-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_zit reindex
	assert_success
	run_zit show +e,t,z,konfig
	assert_success
	assert_output_unsorted - <"$expected"

	run_zit reindex
	assert_success
	run_zit show +e,t,z,konfig
	assert_success
	assert_output_unsorted - <"$expected"
}

function reindex_after_changes { # @test
	run_zit show !md:t
	assert_success
	assert_output - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
	EOM

	cat >md.type <<-EOM
		inline-akte = false
		vim-syntax-type = "test"
	EOM

	run_zit checkin .t
	assert_success
	assert_output - <<-EOM
		[!md @220519ab7c918ccbd73c2d4d73502ab2ec76106662469feea2db8960b5d68217 !toml-type-v1]
	EOM

	function verify() {
		run_zit show -format blob !md+t
		assert_success
		assert_output - <<-EOM
			file-extension = 'md'
			vim-syntax-type = 'markdown'
			inline-akte = false
			vim-syntax-type = "test"
		EOM

		run_zit show -format blob !md:t
		assert_success
		assert_output - <<-EOM
			inline-akte = false
			vim-syntax-type = "test"
		EOM
	}

	verify

	run_zit reindex
	assert_success
	run_zit show +e,t,z,konfig
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[!md @220519ab7c918ccbd73c2d4d73502ab2ec76106662469feea2db8960b5d68217 !toml-type-v1]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[konfig @$(get_konfig_sha) !toml-config-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	verify
}
