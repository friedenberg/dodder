#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

function init_compression { # @test
	run_zit info store-version
	assert_success
	assert_output --regexp '[0-9]+'

	# shellcheck disable=SC2034
	storeVersionCurrent="$output"

	run_zit_init_disable_age

	function output_immutable_config() {
		cat - <<-EOM
			---
			! toml-config-immutable-v1
			---

			public-key = 'zit-repo-public_key-v1.*'
			store-version = $storeVersionCurrent
			repo-type = 'working-copy'
			id = 'test-repo-id'

			\[blob-store]
			compression-type = 'zstd'
			lock-internal-files = false
		EOM
	}

	run_zit info-repo config-immutable
	assert_success
	output_immutable_config | assert_output --regexp -

	run_zit cat-blob "$(get_konfig_sha)"
	assert_success

	sha="$(get_konfig_sha)"
	run zstd --decompress .xdg/data/zit/objects/blobs/"${sha:0:2}"/* --stdout
	assert_success
}

function init_and_reindex { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	set_xdg "$wd"

	run_zit_init_disable_age

	run test -f .xdg/data/zit/config-permanent
	assert_success

	run_zit show -format log :konfig
	assert_success
	assert_output - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v1]
	EOM

	run_zit reindex
	assert_success
	run_zit show :t,konfig
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[konfig @$(get_konfig_sha) !toml-config-v1]
	EOM

	run_zit reindex
	assert_success
	run_zit show :t,konfig
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[konfig @$(get_konfig_sha) !toml-config-v1]
	EOM
}

function init_and_deinit { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	set_xdg "$wd"

	run_zit_init_disable_age

	run test -f .xdg/data/zit/config-permanent
	assert_success

	# run cat .zit/Objekten/Akten/c1/a8ed3cf288dd5d7ccdfd6b9c8052a925bc56be2ec97ed0bb345ab1d961c685
	# assert_output wow
	run_zit show -format log :konfig
	assert_success
	assert_output - <<-EOM
		[konfig @$(get_konfig_sha) !toml-config-v1]
	EOM

	# run_zit deinit
	# assert_success
	# assert_output wow

	# run test ! -f .zit/KonfigAngeboren
	# run ls .zit/
	# assert_success
	# assert_output wow
}

function init_and_with_another_age { # @test
	set_xdg "$BATS_TEST_TMPDIR"
	run_zit_init
	age_id="$(zit info-repo age-encryption)"

	mkdir inner
	pushd inner || exit 1

	set_xdg "$(pwd)"
	run_zit init -yin <(cat_yin) -yang <(cat_yang) -age-identity "$age_id" test-repo-id
	assert_success

	run_zit info-repo age-encryption
	assert_success
	assert_output "$age_id"
}

function init_with_non_xdg { # @test
	run_zit_init -override-xdg-with-cwd test-repo-id
	run tree .zit
	assert_output

	run_zit show +konfig,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[konfig @$(get_konfig_sha) !toml-config-v1]
	EOM
}

function non_repo_failure { # @test
	set_xdg "$BATS_TEST_TMPDIR"
	run_zit show +konfig,t
	assert_failure
	assert_output --partial 'not in a zit directory'
}

function init_and_init { # @test
	set_xdg "$BATS_TEST_TMPDIR"
	run_zit_init -override-xdg-with-cwd test-repo-id
	assert_success

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
	assert_output_unsorted - <<-EOM
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_zit show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_zit init -lock-internal-files=false -override-xdg-with-cwd test-repo-id
	assert_failure
	assert_output --partial ': file exists'

	run_zit show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM

	run_zit show :
	assert_success
	assert_output - <<-EOM
		[one/uno @9e2ec912af5dff2a72300863864fc4da04e81999339d9fac5c7590ba8a3f4e11 !md "wow" tag]
	EOM
}

function init_next { # @test
	run_zit info store-version-next
	assert_success
	assert_output --regexp '[0-9]+'

	# shellcheck disable=SC2034
	storeVersionNext="$output"

	run_zit_init -next -override-xdg-with-cwd test-repo-id
	assert_success

	run_zit info-repo store-version
	assert_success
	assert_output "$storeVersionNext"
}
