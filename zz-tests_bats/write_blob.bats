#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

function write_blob_none { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	run_zit write-blob
	assert_success
	assert_output ''
}

function write_blob_null { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	run_zit write-blob - </dev/null
	assert_success
	assert_output 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 - (checked in)'
}

function write_blob_one_file { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	run_zit write-blob <(echo wow)
	assert_success
	assert_output --partial 'f40cd21f276e47d533371afce1778447e858eb5c9c0c0ed61c65f5c5d57caf63 /dev/fd/'

	run_zit cat-blob "f40cd21f276e47d533371afce1778447e858eb5c9c0c0ed61c65f5c5d57caf63"
	assert_success
	assert_output "$(printf "%s\n" wow)"

	run_zit cat-blob-shas
	assert_success
	assert_output --partial "f40cd21f276e47d533371afce1778447e858eb5c9c0c0ed61c65f5c5d57caf63"
}

function write_blob_one_file_one_stdin { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	run_zit write-blob <(echo wow) - </dev/null
	assert_success
	assert_output --partial 'f40cd21f276e47d533371afce1778447e858eb5c9c0c0ed61c65f5c5d57caf63 /dev/fd/'
	assert_output --partial 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 -'
}
