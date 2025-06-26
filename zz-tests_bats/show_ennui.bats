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

function format_mutter_sha_one { # @test
	run_zit show -format sha one/uno+
	assert_success
	sha="$(echo -n "$output" | head -n1)"

	run_zit show -format mutter-sha one/uno
	assert_success
	assert_output - <<-EOM
		$sha
	EOM
}

function format_mutter_one { # @test
	run_zit show -format mutter one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}

function format_mutter_all { # @test
	run_zit show -format mutter :
	assert_success
	assert_output - <<-EOM
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM
}
