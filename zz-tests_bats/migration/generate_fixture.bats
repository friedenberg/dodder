#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/../common.bash"

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

cmd_def=(
	# -verbose
	-predictable-zettel-ids
)

function generate { # @test
	function run_dodder {
		cmd="$1"
		shift
		#shellcheck disable=SC2068,SC2154
		run timeout --preserve-status "2s" "$DODDER_BIN" "$cmd" ${cmd_dodder_def_no_debug[@]} "$@"
	}

	run_dodder info store-version
	assert_success
	assert_output --regexp '[0-9]+'

	# shellcheck disable=SC2034
	storeVersionCurrent="$output"

	run_dodder_init_disable_age

	run_dodder show :b
	assert_success
	assert_output

	run_dodder last
	assert_success
	assert_output

	run_dodder info store-version
	assert_success
	assert_output "$storeVersionCurrent"

	run_dodder show "${cmd_def[@]}" !md:t :konfig
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[konfig @$(get_konfig_sha) !toml-config-v2]
	EOM

	run_dodder show "${cmd_def[@]}" -format text :konfig
	assert_success
	assert_output - <<-EOM
		---
		! toml-config-v2
		---

		default-blob_store = 'default'

		[defaults]
		type = '!md'
		tags = []

		[file-extensions]
		config = 'konfig'
		organize = 'md'
		repo = 'repo'
		tag = 'tag'
		type = 'type'
		zettel = 'zettel'

		[cli-output]
		print-blob_digests = true
		print-colors = true
		print-empty-blob_digests = false
		print-flush = true
		print-include-description = true
		print-include-types = true
		print-inventory_lists = true
		print-matched-dormant = false
		print-tags-always = true
		print-time = true
		print-unchanged = true

		[cli-output.abbreviations]
		zettel_ids = true
		merkle_ids = true

		[tools]
		merge = ['vimdiff']
	EOM

	run_dodder new "${cmd_def[@]}" -edit=false - <<EOM
---
# wow ok
- tag-1
- tag-2
! md
---

this is the body aiiiiight
EOM

	assert_success
	assert_output_unsorted - <<-EOM
		[one/uno @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder show "${cmd_def[@]}" -format tags one/uno
	assert_success
	assert_output "tag-1, tag-2"

	run_dodder new "${cmd_def[@]}" -edit=false - <<EOM
---
# wow ok again
- tag-3
- tag-4
! md
---

not another one
EOM

	assert_success
	assert_output_unsorted - <<-EOM
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
	EOM

	run_dodder show "${cmd_def[@]}" one/dos
	assert_success
	assert_output - <<-EOM
		[one/dos @blake2b256-z3zpdf6uhqd3tx6nehjtvyjsjqelgyxfjkx46pq04l6qryxz4efs37xhkd !md "wow ok again" tag-3 tag-4]
	EOM

	run_dodder checkout "${cmd_def[@]}" one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
	EOM
	cat >one/uno.zettel <<EOM
---
# wow the first
- tag-3
- tag-4
! md
---

last time
EOM

	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [one/uno.zettel @blake2b256-c5xgv9eyuv6g49mcwqks24gd3dh39w8220l0kl60qxt60rnt60lsc8fqv0 !md "wow ok" tag-1 tag-2]
	EOM

	run_dodder checkin "${cmd_def[@]}" -delete one/uno.zettel
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [one/]
		          deleted [one/uno.zettel]
		[one/uno @blake2b256-9ft3m74l5t2ppwjrvfg3wp380jqj2zfrm6zevxqx34sdethvey0s5vm9gd !md "wow the first" tag-3 tag-4]
	EOM

	run_dodder show "${cmd_def[@]}" -format tags one/uno
	assert_success
	assert_output "tag-3, tag-4"
}
