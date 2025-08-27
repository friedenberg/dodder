#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

function last_after_init { # @test
	run_dodder_init_disable_age

	run_dodder last -format inventory-list-sans-tai
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		\\[!md @$(get_type_blob_sha) .* !toml-type-v1]
		\\[konfig @$(get_konfig_sha) .* !toml-config-v2]
	EOM
}

function last_after_type_mutate { # @test
	run_dodder_init_disable_age

	cat >md.type <<-EOM
		inline-akte = false
		vim-syntax-type = "test"
	EOM

	run_dodder checkin .t
	assert_success
	assert_output - <<-EOM
		[!md @220519ab7c918ccbd73c2d4d73502ab2ec76106662469feea2db8960b5d68217 !toml-type-v1]
	EOM

  # TODO use switch to using the inventory list log
  dir_inventory_list="$("$DODDER_BIN" info-repo dir-blob_stores-0-inventory_lists)"
	run bash -c "find '$dir_inventory_list' -type f | wc -l | tr -d \" \""
	assert_success
  # to support both <v10 separate inventory list blob store, and >=v11 combined inventory list blob store
  [[ "$output" -ge 2 ]]

	run_dodder last -format inventory-list-sans-tai
	assert_success
	assert_output --regexp - <<-EOM
		\\[!md @220519ab7c918ccbd73c2d4d73502ab2ec76106662469feea2db8960b5d68217 .* !toml-type-v1]
	EOM
}

function last_organize { # @test
	run_dodder_init_disable_age

	cat >md.type <<-EOM
		binary = false
		vim-syntax-type = "test"
	EOM

	run_dodder checkin .t
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

	run_dodder last -organize
	assert_success
	assert_output - <<-EOM
		[added @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[added-tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @1c62d833a8ba10d4d272c29b849c4ab2e1e4fed1c6576709940453d5370832cf !toml-type-v1 added-tag]
	EOM
}
