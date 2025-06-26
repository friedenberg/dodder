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

function revert_one_zettel { # @test
	run_dodder revert one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}

function revert_all_zettels { # @test
	run_dodder revert :z
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}

function revert_last { # @test
	run_dodder revert -last
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder last
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}
