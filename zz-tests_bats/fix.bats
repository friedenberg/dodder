#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

function can_update_akte { # @test
	# setup
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_dodder_init_disable_age

	assert_success
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
		[et1]
		[et2]
		[one/uno @036a8e44e472523c0306946f2712f372c234f8a24532e933f1509ae4db0da064 !md "bez" et1 et2]
	EOM

	run_dodder show -format text one/uno
	assert_success
	assert_output "$(cat "$expected")"

	# when
	new_akte="$(mktemp)"
	{
		echo the body but new
	} >"$new_akte"

	run_dodder checkin-blob -new-tags et3 one/uno "$new_akte"
	assert_success
	assert_output - <<-EOM
		[et3]
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

	run_dodder show -format text one/uno
	assert_success
	assert_output "$(cat "$expected")"
}
