#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"
  run_zit_init_workspace

	export BATS_TEST_BODY=true
}

teardown() {
	rm_from_version "$version"
}

function recursive_tags_add_one { # @test
	run_zit checkout tag-3:e
	assert_success
	assert_output - <<-EOM
		      checked out [tag-3.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	cat - >tag-3.tag <<-EOM
		---
		- recurse
		---

	EOM

	run_zit checkin -delete .e
	assert_success
	assert_output - <<-EOM
		[recurse @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 recurse]
		          deleted [tag-3.tag]
	EOM

	run_zit show recurse:e,z
	assert_success
	assert_output_unsorted - <<-EOM
		[recurse @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 recurse]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function recursive_tags_add_one_super_tags { # @test
	run_zit checkout tag-3:e
	assert_success
	assert_output - <<-EOM
		      checked out [tag-3.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	cat - >tag-3.tag <<-EOM
		---
		- recurse
		---

	EOM

	run_zit checkin -delete .e
	assert_success
	assert_output - <<-EOM
		[recurse @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 recurse]
		          deleted [tag-3.tag]
	EOM

	run_zit organize -mode commit-directly <<-EOM
		- [tag-3-sub]
	EOM

	assert_success
	assert_output - <<-EOM
		[tag-3-sub @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show -format tags-path tag-3-sub:e
	assert_success
	assert_output_unsorted - <<-EOM
		tag-3-sub [Paths: [TypeSuper:[tag-3 -> recurse] TypeSelf:[tag-3-sub]], All: [recurse:[TypeSuper:[tag-3 -> recurse]] tag-3:[TypeSuper:[tag-3 -> recurse]] tag-3-sub:[TypeSelf:[tag-3-sub]]]]
	EOM

	run_zit show recurse:e,z
	assert_success
	assert_output_unsorted - <<-EOM
		[recurse @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 recurse]
		[tag-3-sub @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM

	run_zit show -format tags-path recurse:z,e
	assert_success
	assert_output_unsorted - <<-EOM
		one/dos [Paths: [TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse] TypeDirect:[tag-4]], All: [recurse:[TypeIndirect:[tag-3 -> recurse]] tag-3:[TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse]] tag-4:[TypeDirect:[tag-4]]]]
		one/uno [Paths: [TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse] TypeDirect:[tag-4]], All: [recurse:[TypeIndirect:[tag-3 -> recurse]] tag-3:[TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse]] tag-4:[TypeDirect:[tag-4]]]]
		recurse [Paths: [TypeSelf:[recurse]], All: [recurse:[TypeSelf:[recurse]]]]
		tag-3 [Paths: [TypeDirect:[recurse] TypeSelf:[tag-3]], All: [recurse:[TypeDirect:[recurse]] tag-3:[TypeSelf:[tag-3]]]]
		tag-3-sub [Paths: [TypeSuper:[tag-3 -> recurse] TypeSelf:[tag-3-sub]], All: [recurse:[TypeSuper:[tag-3 -> recurse]] tag-3:[TypeSuper:[tag-3 -> recurse]] tag-3-sub:[TypeSelf:[tag-3-sub]]]]
	EOM

	run_zit show -format tags-path recurse:e,z
	assert_success
	assert_output_unsorted - <<-EOM
		one/dos [Paths: [TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse] TypeDirect:[tag-4]], All: [recurse:[TypeIndirect:[tag-3 -> recurse]] tag-3:[TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse]] tag-4:[TypeDirect:[tag-4]]]]
		one/uno [Paths: [TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse] TypeDirect:[tag-4]], All: [recurse:[TypeIndirect:[tag-3 -> recurse]] tag-3:[TypeDirect:[tag-3] TypeIndirect:[tag-3 -> recurse]] tag-4:[TypeDirect:[tag-4]]]]
		recurse [Paths: [TypeSelf:[recurse]], All: [recurse:[TypeSelf:[recurse]]]]
		tag-3 [Paths: [TypeDirect:[recurse] TypeSelf:[tag-3]], All: [recurse:[TypeDirect:[recurse]] tag-3:[TypeSelf:[tag-3]]]]
		tag-3-sub [Paths: [TypeSuper:[tag-3 -> recurse] TypeSelf:[tag-3-sub]], All: [recurse:[TypeSuper:[tag-3 -> recurse]] tag-3:[TypeSuper:[tag-3 -> recurse]] tag-3-sub:[TypeSelf:[tag-3-sub]]]]
	EOM
}

function recursive_tags_with_same_root { # @test
	run_zit organize -mode commit-directly <<-EOM
		- [project-one-crit priority-0_must]
		- [project-one-general]
	EOM

	assert_success
	assert_output - <<-EOM
		[project @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[project-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[priority @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[priority-0_must @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[project-one-crit @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 priority-0_must]
		[project-one-general @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show priority-0_must:e
	assert_success
	assert_output_unsorted - <<-EOM
		[priority-0_must @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[project-one-crit @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 priority-0_must]
	EOM

	run_zit organize -mode commit-directly one/uno <<-EOM
		# project-one-crit, project-one-general
		- [one/uno]
	EOM

	assert_success
	assert_output - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md project-one-crit project-one-general]
	EOM

	run_zit show priority-0_must:z
	assert_success
	assert_output_unsorted - <<-EOM
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md project-one-crit project-one-general]
	EOM
}
