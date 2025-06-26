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

function mergetool_fails_outside_workspace { # @test
	run_zit merge-tool .
	assert_failure
}

function mergetool_none { # @test
	run_zit_init_workspace
	run_zit merge-tool .
	assert_success
	assert_output "nothing to merge"
}

function mergetool_conflict_base {
	run_zit_init_workspace
	run_zit checkout one/dos
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
	EOM

	cat - >one/dos.zettel <<-EOM
		---
		# wow ok again
		- get_this_shit_merged
		- tag-3
		- tag-4
		! txt
		---

		not another one, conflict time
	EOM

	run_zit organize -mode commit-directly one/dos <<-EOM
		---
		! txt2
		---

		# new-etikett-for-all
		- [one/dos  tag-3 tag-4] wow ok again
	EOM
	assert_success
	assert_output_unsorted - <<-EOM
		[!txt2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
		[new-etikett-for-all @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[new-etikett-for @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[new-etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[new @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !txt2 "wow ok again" new-etikett-for-all tag-3 tag-4]
	EOM

	run_zit show -format log new-etikett-for-all:z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[new-etikett-for-all @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !txt2 "wow ok again" new-etikett-for-all tag-3 tag-4]
	EOM

	# TODO add better conflict printing output
	run_zit status one/dos.zettel
	assert_success
	assert_output - <<-EOM
		       conflicted [one/dos.zettel]
	EOM
}

function mergetool_conflict_one_local { # @test
	#TODO-project-2022-zit-collapse_skus
	mergetool_conflict_base

	export BATS_TEST_BODY=true

	# TODO add `-delete` option to `merge-tool`
	run_zit merge-tool -merge-tool "/bin/bash -c 'cat \"\$0\" >\"\$3\"'" .
	assert_success
	assert_output - <<-EOM
		[get_this_shit_merged @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !txt2 "wow ok again" get_this_shit_merged new-etikett-for-all tag-3 tag-4]
		          deleted [one/dos.conflict]
		          deleted [one/dos.zettel]
		          deleted [one/]
	EOM

	run_zit show -format blob one/dos
	assert_success
	assert_output - <<-EOM
		not another one
	EOM

	# run_zit status .
	# assert_success
	# assert_output - <<-EOM
	# 	          changed [one/dos.zettel @9f27ee471da4d09872847d3057ab4fe0d34134b5fef472da37b6f70af483d225 !txt "wow ok again" get_this_shit_merged tag-3 tag-4]
	# EOM

	run_zit last
	assert_success
	assert_output_unsorted - <<-EOM
		[get_this_shit_merged @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !txt2 "wow ok again" get_this_shit_merged new-etikett-for-all tag-3 tag-4]
	EOM
}

function mergetool_conflict_one_remote { # @test
	#TODO-project-2022-zit-collapse_skus
	mergetool_conflict_base

	# TODO add `-delete` option to `merge-tool`
	run_zit merge-tool -merge-tool "/bin/bash -c 'cat \"\$2\" >\"\$3\"'" .
	assert_success
	assert_output - <<-EOM
		[!txt @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
		[get_this_shit_merged @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @9f27ee471da4d09872847d3057ab4fe0d34134b5fef472da37b6f70af483d225 !txt "wow ok again" get_this_shit_merged new-etikett-for-all tag-3 tag-4]
		          deleted [one/dos.conflict]
		          deleted [one/dos.zettel]
		          deleted [one/]
	EOM

	run_zit show -format blob one/dos
	assert_success
	assert_output - <<-EOM
		not another one, conflict time
	EOM

	# run_zit status .
	# assert_success
	# assert_output - <<-EOM
	# 	          changed [one/dos.zettel @9f27ee471da4d09872847d3057ab4fe0d34134b5fef472da37b6f70af483d225 !txt "wow ok again" get_this_shit_merged tag-3 tag-4]
	# EOM

	run_zit last
	assert_success
	assert_output_unsorted - <<-EOM
		[!txt @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
		[get_this_shit_merged @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @9f27ee471da4d09872847d3057ab4fe0d34134b5fef472da37b6f70af483d225 !txt "wow ok again" get_this_shit_merged new-etikett-for-all tag-3 tag-4]
	EOM
}

function mergetool_conflict_one_merged { # @test
	#TODO-project-2022-zit-collapse_skus
	mergetool_conflict_base

	run_zit merge-tool -merge-tool "true" .
	assert_failure
}
