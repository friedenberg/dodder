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

function format_mutter_sha_one { # @test
	run_dodder show -format sha one/uno+
	assert_success
	sha="$(echo -n "$output" | head -n1)"

	run_dodder show -format mutter-sha one/uno
	assert_success
	assert_output - <<-EOM
		$sha
	EOM
}

function format_mutter_one { # @test
	run_dodder show -format mutter one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}

function format_mutter_all { # @test
	run_dodder show -format mutter :
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}
