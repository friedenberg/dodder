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

# bats file_tags=user_story:config

function edit_config_and_change { # @test
	export EDITOR="/bin/bash -c 'echo \"this is the body 2\" > \"\$0\"'"
	run_dodder edit-config
	assert_success
	assert_output - <<-EOM
		[konfig @85eb98a5c8f7ccc354f35b846bb24adc1764e88cb907f63293f6902aa105af58]
	EOM
}

function edit_config_and_dont_change { # @test
	export EDITOR="true"
	run_dodder edit-config
	assert_success
	assert_output ''
}
