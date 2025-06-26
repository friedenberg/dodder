#! /usr/bin/env bats

load "common.bash"

function provides_help_with_no_params { # @test
	run zit
	assert_failure
	assert_output --partial 'No subcommand provided.'
}

function can_initialize_without_age { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	set_xdg "$wd"
	run_zit_init_disable_age
	assert_success

	run test -d .xdg/data/zit/objects
	assert_success
}

function can_initialize_with_age { # @test
	run_zit init \
		-yin <(cat_yin) \
		-yang <(cat_yang) \
		-age-identity generate \
		test-repo-id

	assert_success
	assert_output - <<-EOM
		[!md @b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16 !toml-type-v1]
		[konfig @$(get_konfig_sha) !toml-config-v1]
	EOM

	run test -f .xdg/data/zit/config-permanent

	run_zit info-repo age-encryption
	assert_success
	assert_output
}

function can_new_zettel_file { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	to_add="$(mktemp)"
	{
		echo "---"
		echo "# wow"
		echo "- ok"
		echo "! md"
		echo "---"
	} >"$to_add"

	run_zit new -edit=false "$to_add"
	assert_success
	assert_output - <<-EOM
		[ok @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !md "wow" ok]
	EOM

	run_zit show -format text one/uno:z
	assert_success
	assert_output "$(cat "$to_add")"
}

function can_new_zettel { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	expected="$(mktemp)"
	{
		echo "---"
		echo "# wow"
		echo "- ok"
		echo "! md"
		echo "---"
	} >"$expected"

	run_zit new -edit=false -description wow -tags ok
	assert_success
	assert_output - <<-EOM
		[ok @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !md "wow" ok]
	EOM

	run_zit show -format text one/uno:z
	assert_success
	assert_output "$(cat "$expected")"
}

function can_checkout_and_checkin { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	to_add="$(mktemp)"
	{
		echo "---"
		echo "# wow"
		echo "- ok"
		echo "---"
	} >"$to_add"

	run_zit new -edit=false "$to_add"
	assert_success
	assert_output - <<-EOM
		[ok @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !md "wow" ok]
	EOM

	run_zit checkout one/uno
	assert_success
	# assert_output '       (checked out) [one/uno.zettel @9a638e2b183562da6d3c634d5a3841d64bc337c9cf79f8fffa0d0194659bc564 !md "wow"]'

	{
		echo "---"
		echo "# wow"
		echo "- ok"
		echo "---"
		echo
		echo "content"
	} >"one/uno.zettel"

	run_zit checkin one/uno.zettel
	assert_success
	assert_output - <<-EOM
		[one/uno @434728a410a78f56fc1b5899c3593436e61ab0c731e9072d95e96db290205e53 "wow" ok]
	EOM
}

function can_checkout_via_etiketten { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	to_add="$(mktemp)"
	{
		echo "---"
		echo "# wow"
		echo "- ok"
		echo "---"
	} >"$to_add"

	run_zit new -edit=false "$to_add"
	assert_success
	assert_output - <<-EOM
		[ok @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !md "wow" ok]
	EOM

	run_zit checkout -- ok:z
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !md "wow" ok]
	EOM
}

function can_new_zettel_with_metadata { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	expected="$(mktemp)"
	{
		echo ---
		echo "# bez"
		echo - et1
		echo - et2
		echo ! md
		echo ---
	} >"$expected"

	run_zit new -edit=false -description bez -tags et1,et2
	assert_success
	assert_output_unsorted - <<-EOM
		[et1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[et2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !md "bez" et1 et2]
	EOM
}

function indexes_are_implicitly_correct { # @test
	# setup
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
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

	run_zit new -edit=false "$expected"
	assert_success
	assert_output_unsorted - <<-EOM
		[et1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[et2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @036a8e44e472523c0306946f2712f372c234f8a24532e933f1509ae4db0da064 !md "bez" et1 et2]
	EOM

	{
		echo ---
		echo "# bez"
		echo - et1
		echo ! md
		echo ---
		echo
		echo the body
	} >"$expected"

	mkdir -p one
	cp "$expected" "one/uno.zettel"
	run_zit checkin -delete "one/uno.zettel"
	assert_success
	assert_output - <<-EOM
		[one/uno @036a8e44e472523c0306946f2712f372c234f8a24532e933f1509ae4db0da064 !md "bez" et1]
		          deleted [one/uno.zettel]
		          deleted [one/]
	EOM

	# TODO-P2 fix issue with kennung schwanzen
	# run_zit cat-etiketten-schwanzen
	# assert_success
	# assert_output - <<-EOM
	# EOM

	{
		echo one/uno
	} >"$expected"

	#TODO
	# run_zit cat -gattung hinweis
	# assert_output --partial "$(cat "$expected")"
}

function checkouts_dont_overwrite { # @test
	# setup
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age

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
	assert_output - <<-EOM
		[et1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[et2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @036a8e44e472523c0306946f2712f372c234f8a24532e933f1509ae4db0da064 !md "bez" et1 et2]
	EOM

	run_zit checkout one/uno
	assert_success
	assert_output - <<-EOM
		      checked out [one/uno.zettel @036a8e44e472523c0306946f2712f372c234f8a24532e933f1509ae4db0da064 !md "bez" et1 et2]
	EOM

	run cat one/uno.zettel
	assert_success
	assert_output "$(cat "$expected")"

	{
		echo ---
		echo "# bez"
		echo - et1
		echo - et2
		echo ! md
		echo ---
		echo
		echo the body 2
	} >"$expected"

	cat "$expected" >"one/uno.zettel"

	run_zit checkout one/uno:z
	assert_success
	assert_output - <<-EOM
		          changed [one/uno.zettel @65bdb8b57dfc8b0365a68c71b8a465dd2ff7d26ed07602ffe1a1b39367f42228 !md "bez" et1 et2]
	EOM

	run cat one/uno.zettel
	assert_success
	assert_output "$(cat "$expected")"
}

function invalid_flags_exit_false_cleanly { # @test
	run_zit_init_disable_age
	run_zit new -descriptionx="wow" -edit=false
	assert_failure
	assert_output --regexp - <<-EOM
		flag provided but not defined: -descriptionx
	EOM
}
