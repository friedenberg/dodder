#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	version="v$(dodder info store-version)"
	copy_from_version "$DIR" "$version"
}

teardown() {
	rm_from_version "$version"
}

function debug_options_all() { # @test
	run_dodder info -debug=all
	assert_success

  run test -f cpu.pprof
	assert_success

  run test -f heap.pprof
	assert_success

  run test -f trace
	assert_success
}
