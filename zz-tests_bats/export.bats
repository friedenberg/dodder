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

function basic { # @test
	run_zit export +e,konfig,t,z
	assert_success
	assert_output --regexp - <<-EOM
		\\[!md @$(get_type_blob_sha) .* !toml-type-v1]
		\\[konfig @$(get_konfig_sha) .* !toml-config-v1]
		\\[tag .*]
		\\[tag-1 .*]
		\\[tag-2 .*]
		\\[one/uno .* !md "wow ok" tag-1 tag-2]
		\\[tag-3 .*]
		\\[tag-4 .*]
		\\[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 .* !md "wow ok again" tag-3 tag-4]
		\\[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 .* !md "wow the first" tag-3 tag-4]
	EOM
}
