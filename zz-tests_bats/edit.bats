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

# bats file_tags=user_story:workspace

function edit_and_change_workspace { # @test
	run_zit init-workspace
	assert_success

	export EDITOR="/bin/bash -c 'echo \"this is the body 2\" > \"\$0\"'"
	run_zit edit one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @85eb98a5c8f7ccc354f35b846bb24adc1764e88cb907f63293f6902aa105af58]
	EOM

	run_zit show -format blob one/uno
	assert_success
	assert_output - <<-EOM
		this is the body 2
	EOM
}

function edit_and_dont_change_workspace { # @test
	run_zit init-workspace
	assert_success

	export EDITOR="true"
	run_zit edit one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_zit show -format blob one/uno
	assert_success
	assert_output - <<-EOM
		last time
	EOM
}

# bats file_tags=user_story:noworkspace
# TODO fix no-workspace edits

function edit_and_change_no_workspace { # @test
	export EDITOR="/bin/bash -c 'echo \"this is the body 2\" > \"\$0\"'"
	run_zit edit one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @85eb98a5c8f7ccc354f35b846bb24adc1764e88cb907f63293f6902aa105af58]
	EOM

	run_zit show -format blob one/uno
	assert_success
	assert_output - <<-EOM
		this is the body 2
	EOM
}

function edit_and_dont_change_no_workspace { # @test
	export EDITOR="true"
	run_zit edit one/uno
	assert_success
	assert_output - <<-EOM
	EOM

	run_zit show -format blob one/uno
	assert_success
	assert_output - <<-EOM
		last time
	EOM
}

function edit_and_format_no_workspace { # @test
	# shellcheck disable=SC2317
	function editor() {
		out="$(mktemp)"
		zit format-object "$0" >"$out"
		mv "$out" "$0"
	}

	export -f editor

	# shellcheck disable=SC2016
	export EDITOR='bash -c "editor $0"'

	run_zit edit one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11]
	EOM

	run_zit show -format blob one/uno
	assert_success
	assert_output - <<-EOM
		last time
	EOM
}
