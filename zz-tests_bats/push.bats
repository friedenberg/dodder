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

# bats file_tags=user_story:pull,user_story:repo,user_store:xdg,user_story:remote

function bootstrap_with_content {
	set_xdg "$1"
	run_dodder_init

	{
		echo "---"
		echo "# wow"
		echo "- tag"
		echo "! md"
		echo "---"
		echo
		echo "body"
	} >to_add

	run_dodder new -edit=false to_add
	assert_success
	assert_output - <<-EOM
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_dodder new -edit=false - <<-EOM
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
}

function bootstrap_without_content_xdg {
	mkdir -p them || exit 1
	set_xdg "$(realpath them)"

	pushd them || exit 1
	run_dodder_init
	assert_success
	popd || exit 1
}

function bootstrap_without_content {
	mkdir -p them || exit 1

	pushd them || exit 1
	run_dodder_init -override-xdg-with-cwd test-repo-id-them
	assert_success
	popd || exit 1
}

function bootstrap_archive {
	mkdir -p them || exit 1

	pushd them || exit 1
	run_dodder init \
		-override-xdg-with-cwd \
		-repo-type archive \
		-lock-internal-files=false \
		test-repo-id-them

	run_dodder info-repo type
	assert_success
	assert_output 'archive'

	assert_success
	popd || exit 1
}

function push_history_zettel_typ_etikett_no_conflicts { # @test
	them="them"
	set_xdg "$them"
	run_dodder_init

	function print_their_xdg() (
		set_xdg "$them"
		dodder info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder push /them:k +zettel,typ,etikett

	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 bytes)
		copied Blob 2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 (16 bytes)
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 bytes)
	EOM

	set_xdg "$them"
	run_dodder show +zettel,typ,konfig,etikett,repo
	assert_output_unsorted - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v1]
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

function push_history_zettel_typ_etikett_yes_conflicts { # @test
	skip
	them="them"
	bootstrap_with_content "$them"
	assert_success

	function print_their_xdg() (
		set_xdg "$them"
		dodder info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them

	run_dodder push /them +zettel,typ,etikett

	assert_failure
	assert_output_unsorted - <<-EOM
		       conflicted [one/dos]
		       conflicted [one/uno]
		       conflicted [one/uno]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 bytes)
		copied Blob 2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 (16 bytes)
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 bytes)
		needs merge
	EOM

	run_dodder status .
	assert_output - <<-EOM
		       conflicted [one/dos]
		       conflicted [one/uno]
		        untracked [to_add @05b22ebd6705f9ac35e6e4736371df50b03d0e50f85865861fd1f377c4c76e23]
	EOM

	run_dodder show +zettel,typ,konfig,etikett,repo
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[konfig @d904d322213ed86cdc0eabd58d44f55385f9665280f6c03a01e396f22ba2333b !toml-config-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

function push_history_default { # @test
	bootstrap_without_content_xdg

	function print_their_xdg() (
		set_xdg them
		dodder info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder push /them

	assert_success

	run_dodder show +?z,e,t
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	set_xdg them
	run_dodder show +zettel,typ,konfig,etikett #,repo
	assert_output_unsorted - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v1]
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

function push_history_default_only_blobs { # @test
	bootstrap_without_content_xdg

	function print_their_xdg() (
		set_xdg them
		dodder info xdg
	)

	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type native-dotenv-xdg \
		<(print_their_xdg) \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder push -include-objects=false /them

	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		copied Blob .* \(.* bytes)
		copied Blob .* \(.* bytes)
		copied Blob .* \(.* bytes)
		copied Blob .* \(.* bytes)
		copied Blob .* \(.* bytes)
		copied Blob .* \(.* bytes)
		copied Blob .* \(.* bytes)
	EOM

	run_dodder show +?z,e,t
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	set_xdg them
	run_dodder show +zettel,typ,konfig,etikett,repo
	assert_output_unsorted - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v1]
		[!md @$(get_type_blob_sha) !toml-type-v1]
	EOM
}

function push_default_stdio_local_once { # @test
	bootstrap_without_content
	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type stdio-local \
		them \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-local_path-v0]
	EOM

	export BATS_TEST_BODY=true
	run_dodder push /them
	assert_success
	# TODO-P4 assert output of push

	pushd them || exit 1
	run_dodder show :zettel
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
	popd || exit 1
}

function push_history_default_stdio_local_twice { # @test
	bootstrap_without_content
	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type stdio-local \
		them \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-local_path-v0]
	EOM

	run_dodder push /them :z
	assert_success
	assert_output_unsorted --partial - <<-EOM
		remote: [one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md tag-3 tag-4] wow ok again
		remote: [one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md tag-3 tag-4] wow the first
		remote: copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 bytes)
		remote: copied Blob 2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 (16 bytes)
	EOM

	pushd them || exit 1
	run_dodder show :zettel
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
	popd || exit 1

	run_dodder push /them :z

	assert_success
	assert_output_unsorted - <<-EOM
	EOM
}

function push_history_default_stdio_local_archive_twice { # @test
	skip
	bootstrap_archive
	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type stdio-local \
		them \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @[0-9a-z]+ !toml-repo-local_path-v0]
	EOM

	run_dodder push /them

	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		remote: \[[0-9]+\.[0-9]+ @[0-9a-f]+ !inventory_list-v2]
		remote: copied Blob [0-9a-f]+ \([0-9]+ bytes)
		remote: \[[0-9]+\.[0-9]+ @[0-9a-f]+ !inventory_list-v2]
		remote: copied Blob [0-9a-f]+ \([0-9]+ bytes)
		remote: \[[0-9]+\.[0-9]+ @[0-9a-f]+ !inventory_list-v2]
		remote: copied Blob [0-9a-f]+ \([0-9]+ bytes)
		remote: copied Blob [0-9a-f]+ \([0-9]+ bytes)
		remote: \[[0-9]+\.[0-9]+ @[0-9a-f]+ !inventory_list-v2]
		remote: copied Blob [0-9a-f]+ \([0-9]+ bytes)
	EOM

	pushd them || exit 1
	run_dodder show
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[[0-9]+\.[0-9]+ @[0-9a-f]+ .* !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @[0-9a-f]+ .* !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @[0-9a-f]+ .* !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @[0-9a-f]+ .* !inventory_list-v2]
		\[konfig @d23cb9e6237446e0ff798250c9e82862f29afd997581c9aefdf4916cebd00b90 .* !toml-config-v1]
		\[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 .* !toml-type-v1]
		\[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 .* !md "wow ok again" tag-3 tag-4]
		\[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 .* !md "wow the first" tag-3 tag-4]
		\[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 .* !md "wow ok" tag-1 tag-2]
		\[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
		\[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
		\[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
		\[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
		\[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
	EOM
	popd || exit 1

	run_dodder push /them
	assert_success
	assert_output_unsorted ''
}
