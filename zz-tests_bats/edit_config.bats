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
	export EDITOR="/bin/bash -c 'echo \"# this is the body 2\" >> \"\$0\"'"
	run_dodder edit-config
	assert_success
	assert_output - <<-EOM
		[konfig @blake2b256-38u6g2lzl6gn268umk8r92dpjh4e4dcq69esskuz35v00wt4snrszcpztp !toml-config-v2]
	EOM
}

function edit_config_and_dont_change { # @test
	export EDITOR="true"
	run_dodder edit-config
	assert_success
	assert_output ''
}
