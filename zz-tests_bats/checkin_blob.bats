#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	run_zit_init_disable_age
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

	run_zit new -edit=false "$expected"
	assert_success
	assert_output_unsorted - <<-EOM
		[et1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[et2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @036a8e44e472523c0306946f2712f372c234f8a24532e933f1509ae4db0da064 !md "bez" et1 et2]
	EOM

	run_zit show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"

	# when
	new_blob="$(mktemp)"
	{
		echo the body but new
	} >"$new_blob"

	run_zit checkin-blob -new-tags et3 one/uno "$new_blob"
	assert_success
	assert_output - <<-EOM
		[et3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @a8797107a5f9f8d5e7787e275442499dd48d01e82a153b77590a600702451abd !md "bez" et3]
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

	run_zit show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"
}

function checkin_blob_sha { # @test
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

	run_zit new -edit=false "$expected"
	assert_success
	assert_output_unsorted - <<-EOM
		[et1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[et2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @036a8e44e472523c0306946f2712f372c234f8a24532e933f1509ae4db0da064 !md "bez" et1 et2]
	EOM

	run_zit show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"

	# when
	run_zit write-blob <(echo the body but new)
	assert_success
	assert_output --regexp - <<-EOM
		a8797107a5f9f8d5e7787e275442499dd48d01e82a153b77590a600702451abd
	EOM

	run_zit checkin-blob -new-tags et3 one/uno a8797107a5f9f8d5e7787e275442499dd48d01e82a153b77590a600702451abd
	assert_success
	assert_output - <<-EOM
		[et3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @a8797107a5f9f8d5e7787e275442499dd48d01e82a153b77590a600702451abd !md "bez" et3]
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

	run_zit show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"
}
