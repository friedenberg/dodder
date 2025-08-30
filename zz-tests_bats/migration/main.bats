#! /usr/bin/env bats

setup() {
	load "$BATS_CWD/common.bash"

	# for shellcheck SC2154
	export output

	copy_from_version "$DIR/../"
}

teardown() {
	chflags_and_rm
}

function migration_status_empty { # @test
	run_dodder status
	assert_failure
}

function migration_validate_schwanzen { # @test
	run_dodder show -format log :z,e,t
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[tag-1]
		[tag-2]
		[tag-3]
		[tag-4]
		[tag]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function migration_validate_history { # @test
	run_dodder show -format log +z,e,t
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[tag-1]
		[tag-2]
		[tag-3]
		[tag-4]
		[tag]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}

function migration_reindex { # @test
	run_dodder reindex
  assert_success
  assert_output

	run_dodder show +e,konfig,t,z
  assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[tag-1]
		[tag-2]
		[tag-3]
		[tag-4]
		[tag]
		[konfig @$(get_konfig_sha) !toml-config-v2]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}
