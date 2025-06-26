#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"
}

teardown() {
	chflags_and_rm
}

function add { # @test
	run_zit_init_disable_age

	f=to_add.md
	{
		echo test file
	} >"$f"

	run_zit add \
		-delete \
		-tags zz-inbox-2022-11-14 \
		"$f"

	assert_success
	assert_output - <<-EOM
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to_add" zz-inbox-2022-11-14]
		          deleted [to_add.md]
	EOM

	run_zit show -format text one/uno
	assert_success
	assert_output - <<-EOM
		---
		# to_add
		- zz-inbox-2022-11-14
		! md
		---

		test file
	EOM
}

function add_with_dupe_added { # @test
	run_zit_init_disable_age

	f=to_add.md
	{
		echo test file
	} >"$f"

	f2=to_add2.md
	{
		echo test file
	} >"$f2"

	run_zit add \
		-delete \
		-tags zz-inbox-2022-11-14 \
		"$f" "$f2"

	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [to_add.md]
		          deleted [to_add2.md]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to_add to_add2" zz-inbox-2022-11-14]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show -format text one/uno
	assert_success
	assert_output - <<-EOM
		---
		# to_add to_add2
		- zz-inbox-2022-11-14
		! md
		---

		test file
	EOM
}

function add_not_md { # @test
	run_zit_init_disable_age

	f=to_add.pdf
	{
		echo test file
	} >"$f"

	run_zit add \
		-delete \
		-tags zz-inbox-2022-11-14 \
		-each-blob "bash -c 'basename \$0'" \
		"$f"

	assert_success
	assert_output - <<-EOM
		[!pdf @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !pdf "to_add" zz-inbox-2022-11-14]
		      checked out [one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !pdf "to_add" zz-inbox-2022-11-14
		                   one/uno.pdf]
		uno.pdf
		          deleted [to_add.pdf]
	EOM

	run_zit show -format text one/uno
	assert_success
	assert_output - <<-EOM
		---
		# to_add
		- zz-inbox-2022-11-14
		! pdf
		---

		test file
	EOM
}

function add_1 { # @test
	run_zit_init_disable_age

	f=to_add.md
	{
		echo test file
	} >"$f"

	run_zit add \
		-delete \
		-tags zz-inbox-2022-11-14 \
		"$f"

	assert_success
	assert_output - <<-EOM
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to_add" zz-inbox-2022-11-14]
		          deleted [to_add.md]
	EOM
}

function add_2 { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	f=to_add.md
	{
		echo test file
	} >"$f"

	f2=to_add2.md
	{
		echo test file 2
	} >"$f2"

	run_zit add \
		-delete \
		-tags zz-inbox-2022-11-14 \
		"$f" "$f2"

	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [to_add.md]
		          deleted [to_add2.md]
		[one/dos @6b8e3c36cb01aa01c65ccd86e04935695e9da0580e3dcc8c0c3bce146c274c2c !md "to_add2" zz-inbox-2022-11-14]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to_add" zz-inbox-2022-11-14]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

function add_dot { # @test
	wd="$(mktemp -d)"
	cd "$wd" || exit 1

	run_zit_init_disable_age
	assert_success

	f=to_add.md
	{
		echo test file
	} >"$f"

	f2=to_add2.md
	{
		echo test file 2
	} >"$f2"

	run_zit add \
		-delete \
		-tags zz-inbox-2022-11-14 \
		.

	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [to_add.md]
		          deleted [to_add2.md]
		[one/dos @6b8e3c36cb01aa01c65ccd86e04935695e9da0580e3dcc8c0c3bce146c274c2c !md "to_add2" zz-inbox-2022-11-14]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to_add" zz-inbox-2022-11-14]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

#function add_dedupe_1 { ## @test
#	wd="$(mktemp -d)"
#	cd "$wd" || exit 1

#	run_zit init -disable-age -yin <(cat_yin) -yang <(cat_yang)
#	assert_success

#	f=to_add.md
#	{
#		echo test file
#	} >"$f"

#	run_zit add \
#		-tags zz-inbox-2022-11-14 \
#		"$f"

#	assert_success
#	assert_output - <<-EOM
#		[-zz @48cae50776cad1ddf3e711579e64a1226ae188ddaa195f4eb8cf6d8f32774249]
#		[-zz-inbox @48cae50776cad1ddf3e711579e64a1226ae188ddaa195f4eb8cf6d8f32774249]
#		[-zz-inbox-2022 @48cae50776cad1ddf3e711579e64a1226ae188ddaa195f4eb8cf6d8f32774249]
#		[-zz-inbox-2022-11 @48cae50776cad1ddf3e711579e64a1226ae188ddaa195f4eb8cf6d8f32774249]
#		[-zz-inbox-2022-11-14 @48cae50776cad1ddf3e711579e64a1226ae188ddaa195f4eb8cf6d8f32774249]
#		[one/uno @8f8aa93ce3cb3da0e5eddb2c9556fe37980d0aaf58f2760de451a93ce337b0c2 !md "to_add"]
#		[one/uno @8f8aa93ce3cb3da0e5eddb2c9556fe37980d0aaf58f2760de451a93ce337b0c2 !md "to_add"]
#	EOM

#	run_zit checkout o/u
#	#TODO-P2 fix race condition
#	assert_success
#	assert_output - <<-EOM
#		      checked out [one/uno.zettel @8f8aa93ce3cb3da0e5eddb2c9556fe37980d0aaf58f2760de451a93ce337b0c2 !md "to_add"]
#	EOM

#	{
#		echo '---'
#		echo '# new title'
#		echo '- new-tag'
#		echo '! md'
#		echo '---'
#		echo ''
#		echo 'test file'
#	} >one/uno.zettel

#	run_zit checkin -delete one/uno.zettel
#	assert_success
#	assert_output - <<-EOM
#		[-new @48cae50776cad1ddf3e711579e64a1226ae188ddaa195f4eb8cf6d8f32774249]
#		[-new-tag @48cae50776cad1ddf3e711579e64a1226ae188ddaa195f4eb8cf6d8f32774249]
#		[one/uno @d4853a453015235e41b9513f7e70d91b1a28212f9bd342daf5024b84f35d209f !md "new title"]
#		          deleted [one/uno.zettel]
#		          deleted [one]
#	EOM

#	run zit add \
#		-predictable-zettel-ids \
#		-delete \
#		-tags new-etikett-2 \
#		"$f"

#	run zit show o/u
#	#TODO-P2 fix race condition
#	assert_success
#	assert_output - <<-EOM
#		---
#		# new title
#		- new-etikett-2
#		- new-tag
#		! md
#		---

#		test file
#	EOM
#}

function add_several_with_spaces_in_filename { # @test
	run_zit_init_disable_age

	f="to add.md"
	{
		echo test file
	} >"$f"

	f2="to add2.md"
	{
		echo test file
		echo two!!!!
	} >"$f2"

	run_zit add \
		-delete \
		-tags zz-inbox-2022-11-14 \
		"$f" "$f2"

	assert_success
	assert_output_unsorted - <<-EOM
		          deleted [to add.md]
		          deleted [to add2.md]
		[one/dos @c36af86311166fbaf9cd58f4a161f8dd14618b8242f64ced5b40acd5ed1d1c26 !md "to add2" zz-inbox-2022-11-14]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to add" zz-inbox-2022-11-14]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show -format text one/uno
	assert_success
	assert_output - <<-EOM
		---
		# to add
		- zz-inbox-2022-11-14
		! md
		---

		test file
	EOM
}

function add_each_blob { # @test
	run_zit_init_disable_age

	f="to add.md"
	{
		echo test file
	} >"$f"

	run_zit add \
		-each-blob "cat" \
		-delete \
		-tags zz-inbox-2022-11-14 \
		"$f"

	assert_success
	assert_output_unsorted - <<-EOM
		                   one/uno.md]
		          deleted [to add.md]
		      checked out [one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to add" zz-inbox-2022-11-14
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to add" zz-inbox-2022-11-14]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		test file
	EOM
}

function add_organize { # @test
	run_zit_init_disable_age

	function editor() {
		# shellcheck disable=SC2317
		cp "$1" organize.md
	}

	export -f editor

	# shellcheck disable=SC2016
	export EDITOR='bash -c "editor $0"'

	f="to add.md"
	{
		echo test file
	} >"$f"

	run_zit add \
		-each-blob "cat" \
		-organize \
		-delete \
		-tags zz-inbox-2022-11-14 \
		"$f"

	assert_success
	assert_output - <<-EOM
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-inbox-2022-11-14 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to add" zz-inbox-2022-11-14]
		      checked out [one/uno @55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4 !md "to add" zz-inbox-2022-11-14
		                   one/uno.md]
		test file
		          deleted [to add.md]
	EOM

	run cat organize.md
	assert_success
	assert_output - <<-EOM
		---
		% instructions: to prevent an object from being checked in, delete it entirely
		% delete:true delete once checked in
		- zz-inbox-2022-11-14
		---

		- ["to add.md"]
	EOM
}
