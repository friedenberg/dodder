#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

# bats file_tags=user_story:clone,user_story:repo,user_store:xdg,user_story:remote

function bootstrap {
	mkdir -p "$1"
	pushd "$1" || exit 1
	run_zit_init -override-xdg-with-cwd "test-repo-id-them"

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

	popd || exit 1
}

function print_their_xdg() {
	pushd "$1" >/dev/null || exit 1
	zit info xdg
}

function run_clone_default_with() {
	run_zit clone \
		-age-identity none \
		-yin <(cat_yin) \
		-yang <(cat_yang) \
		-lock-internal-files=false \
		"$@"
}

function try_add_new_after_clone {
	run_zit init-workspace
	assert_success

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

function clone_history_zettel_typ_etikett { # @test
	them="them"
	bootstrap "$them"
	assert_success

	us="us"
	set_xdg "$us"
	run_clone_default_with \
		-remote-type native-dotenv-xdg \
		test-repo-id-us \
		<(print_their_xdg them) \
		+zettel,typ,etikett

	assert_success
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[konfig @b2c9398d2585afe1be26ed36a13703c051311256dc9dab03cf826b377ba237a6 !toml-config-v1]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		copied Blob b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 (51 bytes)
	EOM

	try_add_new_after_clone
}

function clone_history_zettel_typ_etikett_stdio_local { # @test
	them="them"
	bootstrap "$them"
	assert_success

	us="us"
	set_xdg "$us"
	run_clone_default_with \
		-remote-type stdio-local \
		test-repo-id-us \
		"$(realpath them)" \
		+zettel,typ,etikett

	assert_success
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[konfig @b2c9398d2585afe1be26ed36a13703c051311256dc9dab03cf826b377ba237a6 !toml-config-v1]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		copied Blob b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 (51 bytes)
	EOM

	try_add_new_after_clone
}

function clone_history_one_zettel_stdio_local { # @test
	them="them"
	bootstrap "$them"
	assert_success

	us="us"
	set_xdg "$us"
	run_clone_default_with \
		-remote-type stdio-local \
		test-repo-id-us \
		"$(realpath them)" \
		o/d+

	assert_success
	assert_output_unsorted - <<-EOM
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		[konfig @b2c9398d2585afe1be26ed36a13703c051311256dc9dab03cf826b377ba237a6 !toml-config-v1]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
	EOM
}

function clone_history_zettel_typ_etikett_stdio_ssh { # @test
	skip
	them="them"
	bootstrap "$them"
	assert_success

	us="us"
	set_xdg "$us"
	run_clone_default_with \
		-remote-type stdio-local \
		test-repo-id-us \
		"$(realpath them)" \
		+zettel,typ,etikett

	assert_success
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[konfig @b2c9398d2585afe1be26ed36a13703c051311256dc9dab03cf826b377ba237a6 !toml-config-v1]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		copied Blob b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 (51 bytes)
	EOM

	try_add_new_after_clone
}

function clone_history_default_allow_conflicts { # @test
	them="them"
	bootstrap "$them"
	assert_success

	us="us"
	set_xdg "$us"
	run_clone_default_with \
		-remote-type native-dotenv-xdg \
		test-repo-id-us \
		<(print_their_xdg them)

	assert_success

	run_zit show +?z,t,e
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	try_add_new_after_clone
}

function clone_archive_history_default_allow_conflicts { # @test
	them="them"
	bootstrap "$them"
	assert_success

	export BATS_TEST_BODY=true

	us="us"
	set_xdg "$us"
	run_clone_default_with \
		-repo-type archive \
		-remote-type native-dotenv-xdg \
		test-repo-id-us \
		<(print_their_xdg them)

	assert_success
	assert_output ''

	run_zit show
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[[0-9]+\.[0-9]+ @[0-9a-f]+ .* !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @[0-9a-f]+ .* !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @[0-9a-f]+ .* !inventory_list-v2]
		\[konfig @9ad1b8f2538db1acb65265828f4f3d02064d6bef52721ce4cd6d528bc832b822 .* !toml-config-v1]
		\[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 .* !toml-type-v1]
		\[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b .* !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		\[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 .* !md "wow" tag]
		\[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
		\[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
		\[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
	EOM
}

# TODO fix issue with start_server spawning zit processes that do not get cleaned up later
function clone_history_zettel_type_tag_port { # @test
	skip
	them="them"
	bootstrap "$them"
	assert_success

	start_server them

	# shellcheck disable=SC2154
	run echo "$server_PID"
	trap 'kill $server_PID' EXIT
	assert_output 'x'

	us="us"
	set_xdg "$us"
	# shellcheck disable=SC2154
	run_clone_default_with \
		-remote-type url \
		test-repo-id-us \
		"http://localhost:$port" \
		+zettel,typ,etikett

	assert_success
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[konfig @b2c9398d2585afe1be26ed36a13703c051311256dc9dab03cf826b377ba237a6 !toml-config-v1]
		[one/dos @024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_first @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[this_is_the_second @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		copied Blob 024948601ce44cc9ab070b555da4e992f111353b7a9f5569240005639795297b (36 bytes)
		copied Blob 9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 (5 bytes)
		copied Blob b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 (51 bytes)
	EOM

	try_add_new_after_clone
}
