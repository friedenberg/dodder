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
	run_dodder_init -override-xdg-with-cwd "test-repo-id-them"

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
		[one/uno @blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc !md "wow" tag]
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
		[one/dos @blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
	EOM

	popd || exit 1
}

function print_their_xdg() {
	pushd "$1" >/dev/null || exit 1
	"$DODDER_BIN" info xdg
}

function run_clone_default_with() {
	run_dodder clone \
		-encryption none \
		-yin <(cat_yin) \
		-yang <(cat_yang) \
		-lock-internal-files=false \
		"$@"
}

function try_add_new_after_clone {
	run_dodder init-workspace
	assert_success

	run_dodder new -edit=false - <<-EOM
		---
		# zettel after clone description
		! md
		---

		zettel after clone body
	EOM

	assert_success
	assert_output - <<-EOM
		[two/uno @blake2b256-kn7w3q7c3xvfa2p78wny0h79f7hd72nxtded0gvymu33wcnr2qmscl46ar !md "zettel after clone description"]
	EOM
}

function clone_history_zettel_type_tag { # @test
	them="them"
	bootstrap "$them"
	assert_success

  BATS_TEST_FILE=true

	us="us"
	set_xdg "$us"
	run_clone_default_with \
		-remote-type native-dotenv-xdg \
		test-repo-id-us \
		<(print_their_xdg them) \
		+zettel,typ,etikett

	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		\[konfig @blake2b256-.* !toml-config-v2]
		\[one/dos @blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		\[one/uno @blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc !md "wow" tag]
		copied Blob blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 \(36 B)
		copied Blob blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc \(5 B)
		copied Blob blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m \(51 B)
	EOM

	try_add_new_after_clone
}

function clone_history_zettel_type_tag_stdio_local { # @test
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
	assert_output_unsorted --regexp - <<-EOM
		\[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		\[konfig @blake2b256-.+ !toml-config-v2]
		\[one/dos @blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		\[one/uno @blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc !md "wow" tag]
		copied Blob blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 \(36 B)
		copied Blob blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc \(5 B)
		copied Blob blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m \(51 B)
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
	assert_output_unsorted --regexp - <<-'EOM'
		copied Blob blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 \(36 B)
		\[konfig @blake2b256-.* !toml-config-v2]
		\[one/dos @blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
	EOM
}

function clone_history_zettel_type_tag_stdio_ssh { # @test
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
	assert_output_unsorted --regexp - <<-'EOM'
		\[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		\[konfig @blake2b256-.+ !toml-config-v2]
		\[one/dos @blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		\[one/uno @blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc !md "wow" tag]
		copied Blob blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 \(36 B)
		copied Blob blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc \(5 B)
		copied Blob blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m \(51 B)
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

	run_dodder show +?z,t,e
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		[one/dos @blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc !md "wow" tag]
	EOM

	try_add_new_after_clone
}

# TODO fix issue with start_server spawning dodder processes that do not get cleaned up later
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
		[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		[konfig @b2c9398d2585afe1be26ed36a13703c051311256dc9dab03cf826b377ba237a6 !toml-config-v2]
		[one/dos @blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 !md "zettel with multiple etiketten" this_is_the_first this_is_the_second]
		[one/uno @blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc !md "wow" tag]
		[this_is_the_first]
		[this_is_the_second]
		copied Blob blake2b256-fm7kce7793j3npevpm29spk04r6ycxv38dvx3hjxlzl8tcm5m3qq2mml86 (36 B)
		copied Blob blake2b256-gu738nunyrnsqukgqkuaau9zslu0fhwg4dgs9ltuyvnlp42wal8sdpn2hc (5 B)
		copied Blob blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m (51 B)
	EOM

	try_add_new_after_clone
}
