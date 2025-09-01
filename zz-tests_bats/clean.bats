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

function prepare_checkouts() {
	run_dodder_init_workspace
	run_dodder checkout :z,t,e
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [md.type @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

# bats file_tags=user_story:clean

# bats test_tags=user_story:workspace
function clean_fails_outside_workspace { # @test
	run_dodder clean .
	assert_failure
}

# bats file_tags=user_story:workspace

function clean_all { # @test
	prepare_checkouts
	run_dodder clean .
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [md.type]
		          deleted [one/]
		          deleted [one/dos.zettel]
		          deleted [one/uno.zettel]
	EOM

	run_find
	assert_output '.'
}

function clean_zettels { # @test
	prepare_checkouts
	run_dodder clean .z
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [one/dos.zettel]
		          deleted [one/uno.zettel]
		          deleted [one/]
	EOM

	run_find
	assert_success
	assert_output_unsorted - <<-EOM
		.
		./md.type
	EOM
}

function clean_all_dirty_wd { # @test
	prepare_checkouts
	cat >one/uno.zettel <<-EOM
		---
		# wildly different
		- etikett-one
		! md
		---

		newest body
	EOM

	cat >one/dos.zettel <<-EOM
		---
		# dos wildly different
		- etikett-two
		! md
		---

		dos newest body
	EOM

	cat >md.type <<-EOM
		inline-akte = false
		vim-syntax-type = "test"
	EOM

	cat >da-new.type <<-EOM
		inline-akte = true
		vim-syntax-type = "da-new"
	EOM

	cat >zz-archive.tag <<-EOM
		hide = true
	EOM

	run_dodder clean .
	assert_success
	assert_output_unsorted - <<-EOM
	EOM

	run_find
	assert_success
	assert_output_unsorted - <<-EOM
		.
		./md.type
		./one
		./one/uno.zettel
		./one/dos.zettel
		./da-new.type
		./zz-archive.tag
	EOM
}

function clean_all_force_dirty_wd { # @test
	prepare_checkouts
	cat >one/uno.zettel <<-EOM
		---
		# wildly different
		- etikett-one
		! md
		---

		newest body
	EOM

	cat >one/dos.zettel <<-EOM
		---
		# dos wildly different
		- tag-two
		! md
		---

		dos newest body
	EOM

	cat >md.type <<-EOM
		inline-akte = false
		vim-syntax-type = "test"
	EOM

	cat >da-new.type <<-EOM
		inline-akte = true
		vim-syntax-type = "da-new"
	EOM

	cat >zz-archive.tag <<-EOM
		hide = true
	EOM

	run_dodder clean -force .
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [da-new.type]
		          deleted [md.type]
		          deleted [one/dos.zettel]
		          deleted [one/uno.zettel]
		          deleted [one/]
		          deleted [zz-archive.tag]
	EOM

	run_find
	assert_success
	assert_output '.'
}

function clean_hidden { # @test
	prepare_checkouts
	run_dodder show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
	run_dodder organize -mode commit-directly :z <<-EOM
		- [one/uno  !md zz-archive tag-3 tag-4] wow the first
	EOM
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4 zz-archive]
	EOM

	run_dodder dormant-add zz-archive
	assert_success
	assert_output ''

	run_dodder show :z
	assert_success
	assert_output - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
	EOM

	run_dodder show :?z
	assert_success
	assert_output_unsorted - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4 zz-archive]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
	EOM

	run_dodder checkout -force one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4 zz-archive]
	EOM

	run_dodder clean !md.z
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [one/]
		          deleted [one/dos.zettel]
		          deleted [one/uno.zettel]
	EOM
}

function clean_mode_blob_hidden { # @test
	prepare_checkouts
	run_dodder organize -mode commit-directly :z <<-EOM
		- [one/uno  !md zz-archive tag-3 tag-4] wow the first
	EOM
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4 zz-archive]
	EOM

	run_dodder dormant-add zz-archive
	assert_success
	assert_output ''

	run_dodder checkout -force -mode blob one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4 zz-archive
		                   one/uno.md]
	EOM

	run_dodder clean !md.z
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [one/uno.md]
		          deleted [one/dos.zettel]
		          deleted [one/]
	EOM
}

function clean_mode_blob { # @test
	run_dodder_init_workspace
	run_dodder checkout -mode blob one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4
		                   one/uno.md]
	EOM

	run_dodder clean .
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [one/uno.md]
		          deleted [one/]
	EOM
}
