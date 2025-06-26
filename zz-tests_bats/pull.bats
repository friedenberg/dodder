#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

# bats file_tags=user_story:pull,user_story:repo,user_store:xdg,user_story:remote

function bootstrap_xdg {
	set_xdg "$1"
	run_zit_init
	bootstrap_content
}

function bootstrap_repo_at_dir_with_name {
	mkdir -p "$1"
	pushd "$1" || exit 1
	run_zit_init -override-xdg-with-cwd "$1"
	bootstrap_content
	popd || exit 1
}

function bootstrap_content {

	{
		echo "---"
		echo "# wow"
		echo "- tag"
		echo "! md"
		echo "---"
		echo
		echo "body"
	} >to_add

	run_zit new -edit=false to_add
	assert_success
	assert_output - <<-EOM
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_zit new -edit=false - <<-EOM
		---
		# zettel with multiple etiketten
		- this_is_the_first
		- this_is_the_second
		! md
		---

		zettel with multiple etiketten body
	EOM

	assert_success
	assert_output - <<-EOM
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
	EOM

	cat - >task.type <<-EOM
		binary = false
	EOM

	run_zit checkin -delete task.type
	assert_success
	assert_output - <<-EOM
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
		          deleted [task.type]
	EOM
}

function try_add_new_after_pull {
	run_zit new -edit=false - <<-EOM
		---
		# zettel after clone description
		! md
		---

		zettel after clone body
	EOM

	assert_success
	assert_output - <<-EOM
		[two/uno @13af191e86dcd8448565157de81919f19337656787f3d0fdd90b5335d2170f3f !md "zettel after clone description"]
	EOM
}

function pull_history_zettel_typ_etikett_no_conflicts { # @test
	them="them"
	bootstrap_xdg "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		zit info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_zit_init_disable_age

	run_zit remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit pull /them +zettel,typ,etikett

	assert_success
	assert_output_unsorted - <<-EOM
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		copied Blob bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 (15 bytes)
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	try_add_new_after_pull
}

function pull_history_zettel_type_tag_no_conflicts_stdio_local { # @test
	bootstrap_repo_at_dir_with_name them
	assert_success

	set_xdg "$BATS_TEST_TMPDIR"
	export BATS_TEST_BODY=true

	run_zit_init_disable_age

	run_zit remote-add \
		-remote-type stdio-local \
		them \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-local_path-v0]
	EOM

	# TODO make this actually use a socket
	run_zit pull /them +zettel,typ,etikett

	assert_success
	assert_output_unsorted --partial - <<-EOM
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		copied Blob bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 (15 bytes)
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	try_add_new_after_pull
}

function pull_history_zettel_typ_etikett_yes_conflicts_remote_second { # @test
	them="them"
	bootstrap_xdg "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		zit info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"

	run_zit show one/dos+
	assert_success
	assert_output - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
	EOM

	run_zit show +z
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_zit remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit pull /them +zettel,typ,etikett

	assert_failure
	assert_output --partial - <<-EOM
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		       conflicted [one/uno]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		       conflicted [one/dos]
		copied Blob bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 (15 bytes)
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
	EOM

	assert_output --partial - <<-EOM
		needs merge
	EOM

	run_zit status
	assert_success
	assert_output_unsorted - <<-EOM
		       conflicted [one/dos]
		       conflicted [one/uno]
		        untracked [to_add @05b22ebd6705f9ac35e6e4736371df50b03d0e50f85865861fd1f377c4c76e23]
	EOM

	run_zit show +z
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_zit merge-tool -merge-tool "/bin/bash -c 'cat \"\$2\" >\"\$3\"'" .
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		          deleted [one/dos.conflict]
		          deleted [one/uno.conflict]
		          deleted [one/]
	EOM

	# TODO make sure merging includes the REMOTE in addition to the MERGED
	run_zit show +z
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_zit show -format text one/dos
	assert_success
	assert_output - <<-EOM
		---
		# zettel with multiple etiketten
		- this_is_the_first
		- this_is_the_second
		! md
		---

		zettel with multiple etiketten body
	EOM

	run_zit show one/dos+
	assert_success
	assert_output - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
	EOM

	try_add_new_after_pull
}

function pull_history_zettel_typ_etikett_yes_conflicts_allowed_remote_first { # @test
	set_xdg "$BATS_TEST_TMPDIR"
	run_zit_init_disable_age

	run_zit new -edit=false - <<-EOM
		---
		# zettel after clone description
		! md
		---

		zettel after clone body
	EOM

	assert_success
	assert_output - <<-EOM
		[one/uno @13af191e86dcd8448565157de81919f19337656787f3d0fdd90b5335d2170f3f !md "zettel after clone description"]
	EOM

	them="them"
	bootstrap_xdg "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		zit info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_zit remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit pull -allow-merge-conflicts /them +zettel,typ,etikett

	assert_success
	assert_output - <<-EOM
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		copied Blob bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 (15 bytes)
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
	EOM

	run_zit status
	assert_success
	assert_output_unsorted - <<-EOM
		        untracked [to_add @05b22ebd6705f9ac35e6e4736371df50b03d0e50f85865861fd1f377c4c76e23]
	EOM

	run_zit show -format text one/dos
	assert_success
	assert_output - <<-EOM
		---
		# zettel with multiple etiketten
		- this_is_the_first
		- this_is_the_second
		! md
		---

		zettel with multiple etiketten body
	EOM

	run_zit show one/uno+
	assert_success
	assert_output - <<-EOM
		[one/uno @13af191e86dcd8448565157de81919f19337656787f3d0fdd90b5335d2170f3f !md "zettel after clone description"]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM
}

function pull_history_zettel_typ_etikett_yes_conflicts_remote_first { # @test
	set_xdg "$BATS_TEST_TMPDIR"
	run_zit_init_disable_age

	run_zit new -edit=false - <<-EOM
		---
		# zettel after clone description
		! md
		---

		zettel after clone body
	EOM

	assert_success
	assert_output - <<-EOM
		[one/uno @13af191e86dcd8448565157de81919f19337656787f3d0fdd90b5335d2170f3f !md "zettel after clone description"]
	EOM

	them="them"
	bootstrap_xdg "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		zit info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_zit remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit pull /them +zettel,typ,etikett

	assert_failure
	assert_output --partial - <<-EOM
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		       conflicted [one/uno]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		copied Blob bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 (15 bytes)
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
	EOM

	assert_output --partial - <<-EOM
		needs merge
	EOM

	run_zit status
	assert_success
	assert_output_unsorted - <<-EOM
		       conflicted [one/uno]
		        untracked [to_add @05b22ebd6705f9ac35e6e4736371df50b03d0e50f85865861fd1f377c4c76e23]
	EOM

	run_zit merge-tool -merge-tool "/bin/bash -c 'cat \"\$2\" >\"\$3\"'" .
	assert_success
	assert_output - <<-EOM
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		          deleted [one/uno.conflict]
		          deleted [one/]
	EOM

	run_zit show -format text one/dos
	assert_success
	assert_output - <<-EOM
		---
		# zettel with multiple etiketten
		- this_is_the_first
		- this_is_the_second
		! md
		---

		zettel with multiple etiketten body
	EOM

	run_zit show one/uno+
	assert_success
	assert_output - <<-EOM
		[one/uno @13af191e86dcd8448565157de81919f19337656787f3d0fdd90b5335d2170f3f !md "zettel after clone description"]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM
}

function pull_history_default_no_conflict { # @test
	them="them"
	bootstrap_xdg "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		zit info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_zit_init_disable_age

	run_zit remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit pull /them
	assert_success

	run_zit show +?z,t,e
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show one/dos+
	assert_success
	assert_output - <<-EOM
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
	EOM

	run_zit show !md:t
	assert_success
	assert_output - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
	EOM

	run_zit show !task:t
	assert_success
	assert_output - <<-EOM
		[!task @bf2cb7a91cdfdcc84acd1bbaaf0252ff9901977bf76128a578317a42788c4eb6 !toml-type-v1]
	EOM

	try_add_new_after_pull
}

function pull_history_zettel_one_abbr { # @test
	# TODO add support for abbreviations in remote transfers
	skip
	them="them"
	bootstrap_xdg "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		zit info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_zit_init_disable_age

	run_zit remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit pull -include-blobs=false /them o/u+

	assert_success
	assert_output_unsorted - <<-EOM
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_zit show one/uno+
	assert_success
	assert_output - <<-EOM
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM
}

function pull_history_zettels_no_conflict_no_blobs { # @test
	them="them"
	bootstrap_xdg "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		zit info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_zit_init_disable_age

	run_zit remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit pull -include-blobs=false /them +zettel

	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_zit show one/dos+
	assert_success
	assert_output - <<-EOM
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
	EOM

	run_zit show -format blob one/dos
	assert_failure

	try_add_new_after_pull
}
