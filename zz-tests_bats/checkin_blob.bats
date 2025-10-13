#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	run_dodder_init_disable_age
	assert_success

	# for shellcheck SC2154
	export output
}

function checkin_blob_filepath { # @test
	expected="$(mktemp)"
	{
		echo ---
		echo "# bez"
		echo - et1
		echo - et2
		echo ! md
		echo ---
		echo
		echo the body
	} >"$expected"

	run_dodder new -edit=false "$expected"
	assert_success
	assert_output_unsorted - <<-EOM
		[one/uno @blake2b256-vl6ghtv2jsxppshflt86ardlx55ctn8jswx8j59tnv8r99uhs63syxsruy !md "bez" et1 et2]
	EOM

	run_dodder show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"

	# when
	new_blob="$(mktemp)"
	{
		echo the body but new
	} >"$new_blob"

	run_dodder checkin-blob -new-tags et3 one/uno "$new_blob"
	assert_success
	assert_output - <<-EOM
		[one/uno @blake2b256-2qwngrkkpcptsnphu6jcyrwmtpyxux0hmsg4pjfpsn0tr7yt732sgk5lza !md "bez" et3]
	EOM

	# then
	{
		echo ---
		echo "# bez"
		echo - et3
		echo ! md
		echo ---
		echo
		echo the body but new
	} >"$expected"

	run_dodder show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"
}

function checkin_blob_digest { # @test
	expected="$(mktemp)"
	{
		echo ---
		echo "# bez"
		echo - et1
		echo - et2
		echo ! md
		echo ---
		echo
		echo the body
	} >"$expected"

	run_dodder new -edit=false "$expected"
	assert_success
	assert_output_unsorted - <<-EOM
		[one/uno @blake2b256-vl6ghtv2jsxppshflt86ardlx55ctn8jswx8j59tnv8r99uhs63syxsruy !md "bez" et1 et2]
	EOM

	run_dodder show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"

	# when
	run_dodder blob_store-write <(echo the body but new)
	assert_success
	assert_output --regexp - <<-EOM
		blake2b256-2qwngrkkpcptsnphu6jcyrwmtpyxux0hmsg4pjfpsn0tr7yt732sgk5lza
	EOM

	run_dodder checkin-blob -new-tags et3 one/uno blake2b256-2qwngrkkpcptsnphu6jcyrwmtpyxux0hmsg4pjfpsn0tr7yt732sgk5lza
	assert_success
	assert_output - <<-EOM
		[one/uno @blake2b256-2qwngrkkpcptsnphu6jcyrwmtpyxux0hmsg4pjfpsn0tr7yt732sgk5lza !md "bez" et3]
	EOM

	# then
	{
		echo ---
		echo "# bez"
		echo - et3
		echo ! md
		echo ---
		echo
		echo the body but new
	} >"$expected"

	run_dodder show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"
}
