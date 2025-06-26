#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"

	run_zit init-workspace
	assert_success

	run_zit checkout :z,t,e
	assert_success

	export BATS_TEST_BODY=true
}

teardown() {
	rm_from_version "$version"
}

function diff_all_same { # @test
	run_zit diff .
	assert_success
	assert_output_unsorted - <<-EOM
	EOM
}

function diff_all_diff { # @test
	echo wowowow >>one/uno.zettel
	run_zit diff one/uno.zettel
	assert_success
	assert_output - <<-EOM
		--- one/uno:zettel
		+++ one/uno.zettel
		@@ -6,3 +6,4 @@
		 ---
		 
		 last time
		+wowowow
	EOM
}
