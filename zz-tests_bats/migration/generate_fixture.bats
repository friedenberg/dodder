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
	run_zit info store-version
	assert_success
	assert_output --regexp '[0-9]+'

	# shellcheck disable=SC2034
	storeVersionCurrent="$output"

	run_zit_init_disable_age

	run_zit show :b
	assert_success
	assert_output

	run_zit last
	assert_success
	assert_output

	run_zit info store-version
	assert_success
	assert_output "$storeVersionCurrent"

	run_zit show "${cmd_def[@]}" !md:t :konfig
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[konfig @$(get_konfig_sha) !toml-config-v1]
	EOM

	run_zit new "${cmd_def[@]}" -edit=false - <<EOM
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
		[one/uno @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show "${cmd_def[@]}" -format tags one/uno
	assert_success
	assert_output "tag-1, tag-2"

	run_zit new "${cmd_def[@]}" -edit=false - <<EOM
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
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show "${cmd_def[@]}" one/dos
	assert_success
	assert_output - <<-EOM
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
	EOM

	run_zit checkout "${cmd_def[@]}" one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
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
		      checked out [one/uno.zettel @3aa85276929951b03184a038ca0ad67cba78ae626f2e3510426b5a17a56df955 !md "wow ok" tag-1 tag-2]
	EOM

	run_zit checkin "${cmd_def[@]}" -delete one/uno.zettel
	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [one/]
		          deleted [one/uno.zettel]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_zit show "${cmd_def[@]}" -format tags one/uno
	assert_success
	assert_output "tag-3, tag-4"
}
