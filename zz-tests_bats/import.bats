#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	copy_from_version "$DIR"

  skip "TODO add support for blob config files in import"
}

teardown() {
	chflags_and_rm
}

function import { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	set_xdg "$BATS_TEST_TMPDIR"
	run_dodder export -print-time=true +z,e,t
	assert_success
	echo "$output" | zstd >list

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	run_dodder import \
		-inventory-list "$list" \
		-blobs "$blobs" \
		-compression-type zstd
	assert_success

	run_dodder show +z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag]
		[tag-1]
		[tag-2]
		[tag-3]
		[tag-4]
	EOM

	run_dodder show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function import_one_tai_same { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	run_dodder show -format tai one/uno
	tai="$output"

	run_dodder export -print-time=true one/uno [tag ^tag-1 ^tag-2]:e
	assert_success
	echo "$output" | zstd >list

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	echo "$blobs"

	run_dodder import \
		-inventory-list "$list" \
		-blobs "$blobs" \
		-compression-type zstd

	assert_success
	assert_output_unsorted - <<-EOM
		[tag-3]
		[tag-4]
		[tag]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 bytes)
	EOM

	run_dodder show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_dodder show -format tai one/uno
	assert_success
	assert_output "$tai"
}

function import_twice_no_dupes_one_zettel { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	run_dodder show -format inventory-list one/uno+
	assert_success
	echo "$output" | zstd >list

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	run_dodder import -inventory-list "$list" -blobs "$blobs" -compression-type zstd
	assert_success
	assert_output_unsorted - <<-EOM
		[tag-1]
		[tag-2]
		[tag-3]
		[tag-4]
		[tag]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 bytes)
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 bytes)
	EOM

	run_dodder import -inventory-list "$list" -blobs "$blobs" -compression-type zstd
	assert_success
	assert_output - <<-EOM
		           exists [one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok"]
		           exists [one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first"]
	EOM

	run_dodder show :z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[tag-1]
		[tag-2]
		[tag-3]
		[tag-4]
		[tag]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

# TODO add support for conflict resolution
function import_conflict { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	run_dodder export -print-time=true one/uno+
	assert_success
	echo "$output" | zstd >list

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	run_dodder new -edit=false - <<-EOM
		---
		# get out of here!
		- scary
		! md
		---

		ouch a conflict!
	EOM
	assert_success
	assert_output - <<-EOM
		[scary]
		[one/uno @81c3b19e19b4dd2d8e69f413cd253c67c861ec0066e30f90be23ff62fb7b0cf5 !md "get out of here!" scary]
	EOM

	run_dodder import -print-copies=false -inventory-list "$list" -blobs "$blobs" -compression-type zstd
	assert_failure
	assert_output --partial - <<-EOM
		       conflicted [one/uno]
		       conflicted [one/uno]
	EOM

	assert_output --partial - <<-EOM
		needs merge
	EOM
}

function import_twice_no_dupes { # @test
	skip
	(
		mkdir inner
		pushd inner || exit 1
		run_dodder_init
	)

	run_dodder show -format inventory-list +:z,e,t,k
	assert_success
	echo -n "$output" | zstd >list

	list="$(realpath list)"
	blobs="$(realpath .dodder/Objekten2/Akten)"

	pushd inner || exit 1

	run_dodder import -inventory-list "$list" -blobs "$blobs" -compression-type zstd
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[-tag-1]
		[-tag-2]
		[-tag-3]
		[-tag-4]
		[-tag]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 bytes)
		copied Blob 2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 (16 bytes)
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 bytes)
	EOM

	# TODO-P1 fix race condition
	run_dodder import -inventory-list "$list" -blobs "$blobs" -compression-type none
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[-tag-1]
		[-tag-2]
		[-tag-3]
		[-tag-4]
		[-tag]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder show :z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[-tag-1]
		[-tag-2]
		[-tag-3]
		[-tag-4]
		[-tag]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function import_age { # @test
	skip
	run_dodder show -format inventory-list +:z,e,t,k
	assert_success
	echo -n "$output" | zstd >list

	list="$(realpath list)"
	blobs="$(realpath .dodder/Objekten2/Akten)"
	age_id="$(realpath .dodder/AgeIdentity)"

	wd1="$(mktemp -d)"
	cd "$wd1" || exit 1

	run_dodder_init

	run_dodder import -inventory-list "$list" -blobs "$blobs" -age-identity "$(cat "$age_id")" -compression-type zstd
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[-tag-1]
		[-tag-2]
		[-tag-3]
		[-tag-4]
		[-tag]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 bytes)
		copied Blob 2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 (16 bytes)
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 bytes)
	EOM

	run_dodder show one/uno
	assert_success
	assert_output - <<-EOM
		---
		# wow the first
		- tag-3
		- tag-4
		! md
		---

		last time
	EOM
}

function import_inventory_lists { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	set_xdg "$BATS_TEST_TMPDIR"
	run_dodder export -print-time=true
	assert_success
	echo "$output" | zstd >list

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	export BATS_TEST_BODY=true
	run_dodder import \
		-inventory-list "$list" \
		-blobs "$blobs"

	assert_success

	run_dodder show +z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag]
		[tag-1]
		[tag-2]
		[tag-3]
		[tag-4]
	EOM
}
