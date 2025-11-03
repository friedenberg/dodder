#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

function write_blob_none { # @test
	run_dodder_init_disable_age
	assert_success

	run_dodder blob_store-write
	assert_success
	assert_output ''
}

function write_blob_null { # @test
	run_dodder_init_disable_age
	assert_success

	run_dodder blob_store-write - </dev/null
	assert_success
	assert_output 'digest for arg "-" was null'
}

function write_blob_one_file { # @test
	run_dodder_init_disable_age
	assert_success

	run_dodder blob_store-write <(echo wow)
	assert_success
	assert_output --partial 'blake2b256-40mtcwggatwwql4pp9ty93nyugn3r3ppvzs48uza0ze9zltneh3qez5yrs /dev/fd/'

	run_dodder blob_store-cat "blake2b256-40mtcwggatwwql4pp9ty93nyugn3r3ppvzs48uza0ze9zltneh3qez5yrs"
	assert_success
	assert_output "$(printf "%s\n" wow)"

	run_dodder blob_store-cat-ids .default
	assert_success
	assert_output --partial "blake2b256-40mtcwggatwwql4pp9ty93nyugn3r3ppvzs48uza0ze9zltneh3qez5yrs"
}

function write_blob_one_file_one_stdin { # @test
	run_dodder_init_disable_age
	assert_success

	run_dodder blob_store-write <(echo wow) - </dev/null
	assert_success
	assert_output --partial 'blake2b256-40mtcwggatwwql4pp9ty93nyugn3r3ppvzs48uza0ze9zltneh3qez5yrs /dev/fd/'
	assert_output --partial 'digest for arg "-" was null'
}
