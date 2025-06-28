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

# bats file_tags=user_story:init,user_story:workspace,user_story:info

function init_workspace_empty { # @test
	run_dodder info-workspace
	assert_failure
	assert_output --partial - <<-EOM
		not in a workspace
	EOM

	run_dodder init-workspace
	assert_success
	assert_output ''

	run_dodder init-workspace
	assert_failure
	assert_output --partial 'workspace already exists'

	run_dodder info-workspace defaults.type
	assert_success
	assert_output ''

	run_dodder info-workspace defaults.tags
	assert_success
	assert_output '[]'

	run_dodder info-workspace query
	assert_success
	assert_output ''
}

function init_workspace { # @test
	run_dodder info-workspace
	assert_failure
	assert_output --partial - <<-EOM
		not in a workspace
	EOM

	run_dodder init-workspace -query "due" -tags today -type task
	assert_success
	assert_output ''

	run_dodder init-workspace
	assert_failure
	assert_output --partial 'workspace already exists'

	run_dodder info-workspace defaults.type
	assert_success
	assert_output '!task'

	run_dodder info-workspace defaults.tags
	assert_success
	assert_output '[today]'

	run_dodder info-workspace query
	assert_success
	assert_output 'due'
}
