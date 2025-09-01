#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	copy_from_version "$DIR"
}

teardown() {
	chflags_and_rm
}

function basic { # @test
	run_dodder export +e,konfig,t,z
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		---
		! inventory_list-v2
		---

		\[konfig @[0-9a-f]+ .* !toml-config-v2]
		\[!md @[0-9a-f]+ .* !toml-type-v1]
	EOM

	assert_output_unsorted --regexp - <<-'EOM'
		\[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 .* !md "wow ok again" tag-3 tag-4]
	EOM
	assert_output_unsorted --regexp - <<-'EOM'
		\[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 .* !md "wow the first" tag-3 tag-4]
	EOM
	assert_output_unsorted --regexp - <<-'EOM'
		\[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 .* !md "wow ok" tag-1 tag-2]
	EOM
}
