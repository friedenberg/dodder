#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

function fsck_basic { # @test
	run_dodder_init_disable_age

	run_dodder fsck
	assert_success
	assert_output --partial "verification complete"
	assert_output --partial "objects verified:"
	assert_output --partial "objects with errors: 0"
}

function fsck_with_objects { # @test
	run_dodder_init_disable_age

	f=test.md
	{
		echo "test content"
	} >"$f"

	run_dodder add -delete "$f"
	assert_success

	run_dodder fsck
	assert_success
	assert_output --partial "verification complete"
	assert_output --partial "objects with errors: 0"
}

function fsck_skip_probes { # @test
	run_dodder_init_disable_age

	f=test.md
	{
		echo "test content"
	} >"$f"

	run_dodder add -delete "$f"
	assert_success

	run_dodder fsck -skip-probes
	assert_success
	assert_output --partial "verification complete"
	assert_output --partial "objects with errors: 0"
}

function fsck_multiple_objects { # @test
	run_dodder_init_disable_age

	f1=test1.md
	{
		echo "test content 1"
	} >"$f1"

	f2=test2.md
	{
		echo "test content 2"
	} >"$f2"

	f3=test3.md
	{
		echo "test content 3"
	} >"$f3"

	run_dodder add -delete "$f1" "$f2" "$f3"
	assert_success

	run_dodder fsck
	assert_success
	assert_output --partial "verification complete"
	assert_output --partial "objects with errors: 0"
}

function fsck_from_version { # @test
	copy_from_version "$DIR"

	run_dodder fsck
	assert_success
	assert_output --partial "verification complete"
	assert_output --partial "objects with errors: 0"
}

function fsck_from_version_skip_probes { # @test
	copy_from_version "$DIR"

	run_dodder fsck -skip-probes
	assert_success
	assert_output --partial "verification complete"
	assert_output --partial "objects with errors: 0"
}
