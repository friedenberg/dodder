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

function push_history_zettel_type_tag_no_conflicts { # @test
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
		\[/them @blake2b256-.+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder push /them:k +zettel,typ,etikett

	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
		[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
		copied Blob blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd (10 B)
		copied Blob blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd (16 B)
		copied Blob blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 (27 B)
	EOM

	set_xdg "$them"
	run_dodder show +zettel,typ,konfig,etikett,repo
	assert_output_unsorted - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v2]
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
		[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
	EOM
}

function push_history_zettel_type_tag_yes_conflicts { # @test
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
		copied Blob blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd (10 B)
		copied Blob blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd (16 B)
		copied Blob blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 (27 B)
		import failed with conflicts, merging required
	EOM

	run_dodder status .
	assert_output_unsorted - <<-EOM
		       conflicted [one/dos]
		       conflicted [one/uno]
		        untracked [to_add @blake2b256-45lpe4rm9mjvdx8pt04kp5gh04uy77h0m0xtw2fhr0q7vl98g0vqls6hxe]
	EOM

	run_dodder show +zettel,typ,konfig,etikett
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[konfig @$(get_konfig_sha) !toml-config-v2]
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
		[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
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
		\[/them @blake2b256-.+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder push /them

	assert_success

	run_dodder show +?z,e,t
	assert_output_unsorted - <<-EOM
		[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
		[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
	EOM

	set_xdg them
	run_dodder show +zettel,typ,konfig,etikett #,repo
	assert_output_unsorted - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v2]
		[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
		[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
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
		\[/them @blake2b256-.+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder push -include-objects=false /them

	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		copied Blob .* \(.*)
		copied Blob .* \(.*)
		copied Blob .* \(.*)
		copied Blob .* \(.*)
		copied Blob .* \(.*)
		copied Blob .* \(.*)
		copied Blob .* \(.*)
	EOM

	run_dodder show +?z,e,t
	assert_output_unsorted - <<-EOM
		[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
		[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
	EOM

	set_xdg them
	run_dodder show +zettel,typ,konfig,etikett,repo
	assert_output_unsorted - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v2]
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
		\[/them @blake2b256-.+ !toml-repo-local_path-v0]
	EOM

	export BATS_TEST_BODY=true
	run_dodder push /them
	assert_success
	# TODO-P4 assert output of push

	pushd them || exit 1
	run_dodder show :zettel
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
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
		\[/them @blake2b256-.+ !toml-repo-local_path-v0]
	EOM

	run_dodder push /them :z
	assert_success
	assert_output_unsorted --partial - <<-EOM
		(remote) [one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md tag-3 tag-4] wow ok again
		(remote) [one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md tag-3 tag-4] wow the first
		(remote) copied Blob blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd (10 B)
		(remote) copied Blob blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd (16 B)
	EOM

	pushd them || exit 1
	run_dodder show :zettel
	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
	EOM
	popd || exit 1

	run_dodder push /them :z

	assert_success
	assert_output_unsorted - <<-EOM
	EOM
}

function push_history_default_stdio_twice { # @test
	bootstrap_without_content
	set_xdg "$BATS_TEST_TMPDIR"

	run_dodder remote-add \
		-remote-type stdio-local \
		them \
		them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/them @blake2b256-.+ !toml-repo-local_path-v0]
	EOM

	run_dodder push /them

	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\(remote) \[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\(remote) \[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\(remote) \[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\(remote) \[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\(remote) copied Blob blake2b256-.+ \(.*)
		\(remote) copied Blob blake2b256-.+ \(.*)
		\(remote) copied Blob blake2b256-.+ \(.*)
		\(remote) copied Blob blake2b256-.+ \(.*)
		\(remote) copied Blob blake2b256-.+ \(.*)
	EOM

	pushd them || exit 1
	run_dodder show +z,e,t,b
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\[[0-9]+\.[0-9]+ @blake2b256-.+ !inventory_list-v2]
		\[!md @blake2b256-3kj7xgch6rjkq64aa36pnjtn9mdnl89k8pdhtlh33cjfpzy8ek4qnufx0m !toml-type-v1]
		\[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
		\[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
		\[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
	EOM
	popd || exit 1

	run_dodder push /them
	assert_success
	assert_output_unsorted ''
}
