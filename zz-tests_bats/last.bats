#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	# version="v$(zit info store-version)"
	# copy_from_version "$DIR" "$version"
}

teardown() {
	chflags_and_rm
}

function last_after_init { # @test
	run_zit_init_disable_age

	run_zit last -format inventory-list-sans-tai
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		\\[!md @$(get_type_blob_sha) .* !toml-type-v1]
		\\[konfig @$(get_konfig_sha) .* !toml-config-v1]
	EOM
}

function last_after_typ_mutate { # @test
	run_zit_init_disable_age

	cat >md.type <<-EOM
		inline-akte = false
		vim-syntax-type = "test"
	EOM

	run_zit checkin .t
	assert_success
	assert_output - <<-EOM
		[!md @220519ab7c918ccbd73c2d4d73502ab2ec76106662469feea2db8960b5d68217 !toml-type-v1]
	EOM

	run bash -c 'find .xdg/data/zit/objects/inventory_lists -type f | wc -l | tr -d " "'
	assert_success
	assert_output '2'

	run_zit last -format inventory-list-sans-tai
	assert_success
	assert_output --regexp - <<-EOM
		\\[!md @220519ab7c918ccbd73c2d4d73502ab2ec76106662469feea2db8960b5d68217 .* !toml-type-v1]
	EOM
}

function last_organize { # @test
	run_zit_init_disable_age

	cat >md.type <<-EOM
		binary = false
		vim-syntax-type = "test"
	EOM

	run_zit checkin .t
	assert_success
	assert_output - <<-EOM
		[!md @1c62d833a8ba10d4d272c29b849c4ab2e1e4fed1c6576709940453d5370832cf !toml-type-v1]
	EOM

	function editor() {
		# shellcheck disable=SC2317
		cat - >"$1" <<-EOM
			- [!md !toml-type-v1 added-tag]
		EOM
	}

	export -f editor

	# shellcheck disable=SC2016
	export EDITOR='bash -c "editor $0"'

	run_zit last -organize
	assert_success
	assert_output - <<-EOM
		[added @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[added-tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @1c62d833a8ba10d4d272c29b849c4ab2e1e4fed1c6576709940453d5370832cf !toml-type-v1 added-tag]
	EOM
}
