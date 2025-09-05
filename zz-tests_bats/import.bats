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

function import { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	set_xdg "$BATS_TEST_TMPDIR"
	run_dodder info-repo pubkey
	assert_success
	# old_pubkey="$output"

	run_dodder export -print-time=true +z,e,t
	assert_success
	echo "$output" | zstd >list

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	run_dodder info-repo pubkey
	assert_success
	new_pubkey="$output"

	run_dodder import \
		-inventory-list "$list" \
		-blobs "$blobs" \
		-compression-type zstd
	assert_success

	run_dodder show -format inventory_list +z,e,t
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		\\[!md @$(get_type_blob_sha) .* !toml-type-v1]
		\\[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 .* $new_pubkey .* !md "wow ok again" tag-3 tag-4]
		\\[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 .* $new_pubkey .* !md "wow the first" tag-3 tag-4]
		\\[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 .* $new_pubkey .* !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function import_with_overwrite_sig { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	set_xdg "$BATS_TEST_TMPDIR"
	run_dodder info-repo pubkey
	assert_success
	# old_pubkey="$output"

	# run_dodder export -print-time=true +z,e,t
	zstd >list <<-EOM
		---
		! inventory_list-v2
		---

		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 2135591162.342034946 dodder-repo-public_key-v1@ed25519_pub-vhhh5p6qfc9q5fpqm2xmjmetgnagmjpxxqlwlac4uvrhrvjvgevsv5z5q6 dodder-object-sig-v1@ed25519_sig-anhgqrkdqnn6uzvcaj93hr7epr72v8vefv0gkrhd7ktskl6pez2cr8kwe3krrndw8lefh8a7k5dzhete4pjk72zfp4vgf8f0srpksqsy6nn8g !toml-type-v1]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 2135591162.520209927 dodder-repo-public_key-v1@ed25519_pub-vhhh5p6qfc9q5fpqm2xmjmetgnagmjpxxqlwlac4uvrhrvjvgevsv5z5q6 dodder-object-sig-v1@ed25519_sig-jr7jqjh6rq0zd42n03z5vcl2grqr3eg9eqwnuwxj809h3eaxqw58mm3garf4nzenptmu9mhamhtlt9uuxsrt5wl4dshsfsnak3zvgrcelwkhr !md "wow ok" tag-1 tag-2]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 2135591162.606407248 dodder-repo-public_key-v1@ed25519_pub-vhhh5p6qfc9q5fpqm2xmjmetgnagmjpxxqlwlac4uvrhrvjvgevsv5z5q6 dodder-object-sig-v1@ed25519_sig-3ya9fl5nlx7e77qk4vvx2ae7cez8uagywym8f2h5r6f4ern2fhslgtvqjge6fzxjwkkgfr9qjpt0kjjq6slzrm7phraq9jm4z42q2qqnnh2eu !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 2135591162.697539117 dodder-repo-public_key-v1@ed25519_pub-vhhh5p6qfc9q5fpqm2xmjmetgnagmjpxxqlwlac4uvrhrvjvgevsv5z5q6 dodder-object-mother-sig-v1@ed25519_sig-jr7jqjh6rq0zd42n03z5vcl2grqr3eg9eqwnuwxj809h3eaxqw58mm3garf4nzenptmu9mhamhtlt9uuxsrt5wl4dshsfsnak3zvgrcelwkhr dodder-object-sig-v1@ed25519_sig-3ngs79lfywr6ewtdze0c9d3mwk824mymu8xjavzn3uc5s26fzwdy6mz487yasxhd2nqwefjuq3rtnfsj6a4p2u4dcj0wt2h4s2yl6qgm73qt6 !md "wow the first" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 2135591162.697539118 dodder-repo-public_key-v1@ed25519_pub-vhhh5p6qfc9q5fpqm2xmjmetgnagmjpxxqlwlac4uvrhrvjvgevsv5z5q6 dodder-object-mother-sig-v1@ed25519_sig-jr7jqjh6rq0zd42n03z5vcl2grqr3eg9eqwnuwxj809h3eaxqw58mm3garf4nzenptmu9mhamhtlt9uuxsrt5wl4dshsfsnak3zvgrcelwkhr dodder-object-sig-v1@ed25519_sig-3ngs79lfywr6ewtdze0c9d3mwk824mymu8xjavzn3uc5s26fzwdy6mz487yasxhd2nqwefjuq3rtnfsj6a4p2u4dcj0wt2h4s2yl6qgm73qt6 !md "wow the first" tag-3 tag-4]
	EOM

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	run_dodder info-repo pubkey
	assert_success
	new_pubkey="$output"

	run_dodder import \
		-overwrite-signatures=true \
		-inventory-list "$list" \
		-blobs "$blobs" \
		-compression-type zstd
	assert_success

	run_dodder show -format inventory_list +z,e,t
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		\\[!md @$(get_type_blob_sha) .* !toml-type-v1]
		\\[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 .* $new_pubkey .* !md "wow ok again" tag-3 tag-4]
		\\[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 .* $new_pubkey .* !md "wow the first" tag-3 tag-4]
		\\[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 .* $new_pubkey .* !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function import_with_dupes_in_list { # @test
	(
		mkdir inner
		pushd inner || exit 1
		set_xdg "$(pwd)"
		run_dodder_init
	)

	set_xdg "$BATS_TEST_TMPDIR"
	run_dodder info-repo pubkey
	assert_success
	# old_pubkey="$output"

	# run_dodder export -print-time=true +z,e,t
	zstd >list <<-EOM
		---
		! inventory_list-v2
		---

		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 2135591162.342034946 !toml-type-v1]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 2135591162.520209927 !md "wow ok" tag-1 tag-2]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 2135591162.606407248 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 2135591162.697539117 !md "wow the first" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 2135591162.697539118 !md "wow the first" tag-3 tag-4]
	EOM

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	run_dodder info-repo pubkey
	assert_success
	new_pubkey="$output"

	run_dodder import \
		-overwrite-signatures=true \
		-inventory-list "$list" \
		-blobs "$blobs" \
		-compression-type zstd
	assert_success
	assert_output - <<-EOM
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 B)
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		copied Blob 2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 (16 B)
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 B)
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_dodder show -format inventory_list +z,e,t
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		\\[!md @$(get_type_blob_sha) .* !toml-type-v1]
		\\[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 .* $new_pubkey .* !md "wow ok again" tag-3 tag-4]
		\\[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 .* $new_pubkey .* !md "wow the first" tag-3 tag-4]
		\\[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 .* $new_pubkey .* !md "wow ok" tag-1 tag-2]
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
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 B)
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

	run_dodder export -print-time=true one/uno+
	assert_success
	echo "$output" | zstd >list

	list="$(realpath list)"
	blobs="$("$DODDER_BIN" info-repo dir-blob_stores-0-blobs)"

	pushd inner || exit 1
	set_xdg "$(pwd)"

	run_dodder import -inventory-list "$list" -blobs "$blobs" -compression-type zstd
	assert_success
	assert_output_unsorted - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 B)
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 B)
	EOM

	run_dodder import -inventory-list "$list" -blobs "$blobs" -compression-type zstd
	assert_success
	assert_output - <<-EOM
	EOM

	run_dodder show :z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
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
		[one/uno @81c3b19e19b4dd2d8e69f413cd253c67c861ec0066e30f90be23ff62fb7b0cf5 !md "get out of here!" scary]
	EOM

	run_dodder import -print-copies=false -inventory-list "$list" -blobs "$blobs" -compression-type zstd
	assert_failure
	assert_output --partial - <<-EOM
		       conflicted [one/uno]
		       conflicted [one/uno]
	EOM

	assert_output --partial - <<-EOM
		import failed with conflicts, merging required
	EOM
}

function import_twice_no_dupes { # @test
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
	assert_output_unsorted - <<-EOM
		copied Blob 11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 (10 B)
		copied Blob 2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 (16 B)
		copied Blob 3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 (27 B)
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder show +z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder import \
		-inventory-list "$list" \
		-blobs "$blobs" \
		-compression-type zstd
	assert_success
	assert_output_unsorted - <<-EOM
	EOM

	run_dodder show :z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_dodder show +z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
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
	EOM
}
