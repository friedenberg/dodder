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

function workspace_show { # @test
	run_zit init-workspace -query tag-3
	assert_success

	run_zit show
	assert_success
	assert_output_unsorted - <<-eom
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
	eom

	run_zit show :e
	assert_success
	assert_output_unsorted - <<-eom
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	eom

	run_zit show one/uno
	assert_success
	assert_output - <<-eom
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	eom
}

function workspace_edit { # @test
	run_zit init-workspace -query tag-3
	assert_success

	export EDITOR="true"
	run_zit edit
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [tag-3.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_zit show -format blob one/uno
	assert_success
	assert_output - <<-EOM
		last time
	EOM
}

function workspace_checkout { # @test
	run_zit init-workspace -tags tag-3
	assert_success

	run_zit checkout
	assert_success
	assert_output ''

	run_zit checkout :
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_zit show -format blob one/uno.zettel
	assert_success
	assert_output - <<-EOM
		last time
	EOM
}

function workspace_organize { # @test
	run_zit init-workspace -tags tag-3 -query tag-3
	assert_success

	run_zit organize -mode output-only
	assert_success
	assert_output - <<-EOM
		---
		- tag-3
		---
	EOM

	run_zit organize -mode output-only :
	assert_success
	assert_output - <<-EOM
		---
		- tag-3
		---

		- [one/dos !md tag-4] wow ok again
		- [one/uno !md tag-4] wow the first
	EOM

	run_zit organize -mode output-only one/uno
	assert_success
	assert_output - <<-EOM
		---
		- tag-3
		---

		- [one/uno !md tag-4] wow the first
	EOM
}

function workspace_add_no_organize { # @test
	run_zit init-workspace -tags tag-3 -query tag-3
	assert_success

	echo "file to be added" >todo.wow.md

	run_zit add -delete -tags new_tags -description "added file" todo.wow.md
	assert_success
	assert_output - <<-EOM
		[new_tags @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[two/uno @84b683398cc5974fa1e383573fb104d31312c20f6053ef422463f3522e15be15 !md "added file" new_tags tag-3]
		          deleted [todo.wow.md]
	EOM
}

function workspace_add_yes_organize { # @test
	run_zit init-workspace -tags tag-3 -query tag-3
	assert_success

	echo "file to be added1" >1.md
	echo "file to be added2" >2.md

	function editor() {
		# shellcheck disable=SC2317
		cat - >"$1" <<-EOM
			# tag-two

			- [1.md]

			# tag-one

			- [2.md]
		EOM
	}

	export -f editor

	# shellcheck disable=SC2016
	export EDITOR='bash -c "editor $0"'

	run_zit add -organize -delete ./*.md
	assert_success
	assert_output - <<-EOM
		[tag-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[two/uno @38dfdd64dc162365079f6e2b02942ada29fba3aa7cd36cd5e6b13c0fde3777d5 !md "1" tag-two]
		[tag-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/tres @626e7fcba179d01d0d58237102d25aa566b249a09a9e6ed8a5948dacf2d45ead !md "2" tag-one]
		          deleted [1.md]
		          deleted [2.md]
	EOM
}

function workspace_add_yes_organize_omit_one { # @test
	run_zit init-workspace -tags tag-3 -query tag-3
	assert_success

	echo "file to be added1" >1.md
	echo "file to be added2" >2.md

	function editor() {
		# shellcheck disable=SC2317
		cat - >"$1" <<-EOM
			# tag-two

			- [1.md]
		EOM
	}

	export -f editor

	# shellcheck disable=SC2016
	export EDITOR='bash -c "editor $0"'

	run_zit add -organize -delete ./*.md
	assert_success
	assert_output - <<-EOM
		[tag-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[two/uno @38dfdd64dc162365079f6e2b02942ada29fba3aa7cd36cd5e6b13c0fde3777d5 !md "1" tag-two]
		          deleted [1.md]
	EOM
}

function workspace_parent_directory { # @test
	run_zit init-workspace -tags tag-3 -query tag-3
	assert_success

  mkdir -p child
  pushd child || exit 1

  run_zit info-workspace
  assert_success
  assert_output ''
}
