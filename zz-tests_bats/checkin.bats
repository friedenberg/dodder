#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"
	run_zit_init_workspace

	run_zit checkout :z,t,e
	assert_success
	assert_output_unsorted - <<-EOM
		      checked out [md.type @$(get_type_blob_sha) !toml-type-v1]
		      checked out [one/dos.zettel @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		      checked out [one/uno.zettel @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
		      checked out [tag-1.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag-2.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag-3.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag-4.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		      checked out [tag.tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run ls
	assert_success
	assert_output_unsorted - <<-EOM
		md.type
		one
		tag-1.tag
		tag-2.tag
		tag-3.tag
		tag-4.tag
		tag.tag
	EOM

	cat >one/uno.zettel <<-EOM
		---
		# wildly different
		- etikett-one
		! md
		---

		newest body
	EOM

	cat >one/dos.zettel <<-EOM
		---
		# dos wildly different
		- etikett-two
		! md
		---

		dos newest body
	EOM

	cat >md.type <<-EOM
		binary = true
		vim-syntax-type = "test"
	EOM

	cat >zz-archive.tag <<-EOM
		hide = true
	EOM

	export BATS_TEST_BODY=true
}

teardown() {
	rm_from_version "$version"
}

function dirty_one_virtual() {
	cat >one/uno.zettel <<-EOM
		---
		# wildly different
		- etikett-one
		- %virtual
		! md
		---

		newest body
	EOM
}

function checkin_simple_one_zettel { # @test
	run_zit checkin one/uno.zettel
	assert_success
	assert_output - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
	EOM
}

function checkin_two_zettel_hidden { # @test
	run_zit dormant-add etikett-one tag-3
	assert_success

	run_zit checkin .z
	assert_success
	assert_output_unsorted - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
		[etikett-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @b5c4fbaac3b71657edee74de4b947f13dfa104715feb8bab7cfa4dd47cafa3db !md "dos wildly different" etikett-two]
	EOM
}

function checkin_simple_one_zettel_virtual_etikett { # @test
	dirty_one_virtual
	run_zit checkin one/uno.zettel
	assert_success
	assert_output - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" %virtual etikett-one]
	EOM

	run_zit show one/uno
	assert_success
	assert_output - <<-EOM
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
	EOM
}

function checkin_complex_zettel_etikett_negation { # @test
	run_zit checkin ^etikett-two.z
	assert_success
	assert_output - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
	EOM
}

function checkin_simple_all { # @test
	run_zit checkin .
	assert_success
	assert_output_unsorted - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
		[one/dos @b5c4fbaac3b71657edee74de4b947f13dfa104715feb8bab7cfa4dd47cafa3db !md "dos wildly different" etikett-two]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
		[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show -format log :?z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
		[one/dos @b5c4fbaac3b71657edee74de4b947f13dfa104715feb8bab7cfa4dd47cafa3db !md "dos wildly different" etikett-two]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

function checkin_simple_all_dry_run { # @test
	run_zit checkin -dry-run .
	assert_success
	assert_output_unsorted - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
		[one/dos @b5c4fbaac3b71657edee74de4b947f13dfa104715feb8bab7cfa4dd47cafa3db !md "dos wildly different" etikett-two]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
		[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM

	run_zit show -format log :z,e,t
	assert_success
	assert_output_unsorted - <<-EOM
		[!md @$(get_type_blob_sha) !toml-type-v1]
		[tag-1 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-2 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-3 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag-4 @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/dos @2d36c504bb5f4c6cc804c63c983174a36303e1e15a3a2120481545eec6cc5f24 !md "wow ok again" tag-3 tag-4]
		[one/uno @11e1c0499579c9a892263b5678e1dfc985c8643b2d7a0ebddcf4bd0e0288bc11 !md "wow the first" tag-3 tag-4]
	EOM
}

function checkin_simple_typ { # @test
	run_zit checkin .t
	assert_success
	assert_output - <<-EOM
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
	EOM

	run_zit show -format blob !md:t
	assert_success
	assert_output - <<-EOM
		binary = true
		vim-syntax-type = "test"
	EOM

	run_zit last -format box-archive
	assert_success
	assert_output --regexp - <<-'EOM'
		\[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b .* !toml-type-v1]
	EOM

	run_zit show !md:t
	assert_success
	assert_output - <<-EOM
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
	EOM

	run_zit show -format type.vim-syntax-type !md:typ
	assert_success
	assert_output 'toml'

	run_zit show -format type.vim-syntax-type one/uno
	assert_success
	assert_output 'test'
}

function checkin_simple_etikett { # @test
	run_zit checkin zz-archive.tag
	# run_zit checkin zz-archive.e
	assert_success
	assert_output - <<-EOM
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93]
	EOM

	run_zit last -format inventory-list-sans-tai
	assert_success
	assert_output --regexp - <<-'EOM'
		\[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 .*]
		\[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93 .*]
	EOM

	run_zit show -format blob zz-archive?e
	assert_success
	assert_output - <<-EOM
		hide = true
	EOM
}

function checkin_zettel_typ_has_commit_hook { # @test
	cat >typ_with_hook.type <<-EOM
		hooks = """
		return {
		  on_new = function (kinder)
		    kinder["Etiketten"]["on_new"] = true
		    return nil
		  end,
		  on_pre_commit = function (kinder, mutter)
		    kinder["Etiketten"]["on_pre_commit"] = true
		    return nil
		  end,
		}
		"""
	EOM

	run_zit checkin -delete typ_with_hook.type
	assert_success
	assert_output - <<-EOM
		[!typ_with_hook @1f6b9061059a83822901612bc050dd7d966bb5a2ceb917549ca3881728854477 !toml-type-v1]
		          deleted [typ_with_hook.type]
	EOM

	run_zit new -edit=false - <<-EOM
		---
		# test lua
		! typ_with_hook
		---

		should add new etikett
	EOM
	assert_success
	assert_output - <<-EOM
		[on_new @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[on_pre_commit @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[two/uno @edf7b6df934442ad0d6ac9fe4132c5e588391eb307fbbdc3ab6de780e17245a5 !typ_with_hook "test lua" on_new on_pre_commit]
	EOM
}

function checkin_zettel_with_komment { # @test
	run_zit checkin -print-inventory_list=true -comment "message" one/uno.zettel
	assert_success
	assert_output --regexp - <<-'EOM'
		\[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855\]
		\[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855\]
		\[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one\]
		\[[0-9]+\.[0-9]+ @[0-9a-f]{64} !inventory_list-v2 "message"\]
	EOM
}

function checkin_via_organize { # @test
	export EDITOR="true"
	run_zit checkin -organize one/uno.zettel
	assert_success
	assert_output - <<-'EOM'
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:external_ids
function checkin_dot_untracked_fs_blob() { # @test
	cat >test.md <<-EOM
		newest body
	EOM

	run_zit checkin .
	assert_success
	assert_output_unsorted - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
		[one/dos @b5c4fbaac3b71657edee74de4b947f13dfa104715feb8bab7cfa4dd47cafa3db !md "dos wildly different" etikett-two]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "test"]
		[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:external_ids
function checkin_explicit_untracked_fs_blob() { # @test
	cat >test.md <<-EOM
		newest body
	EOM

	run_zit checkin test.md
	assert_success
	assert_output - <<-EOM
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "test"]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:organize, user_story:editor, user_story:external_ids
function checkin_dot_organize_exclude_untracked_fs_blob() { # @test
	cat >test.md <<-EOM
		newest body
	EOM

	export EDITOR="true"
	run_zit checkin -organize .
	assert_success
	assert_output_unsorted - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
		[one/dos @b5c4fbaac3b71657edee74de4b947f13dfa104715feb8bab7cfa4dd47cafa3db !md "dos wildly different" etikett-two]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "test"]
		[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:organize, user_story:editor, user_story:external_ids
function checkin_explicit_organize_include_untracked_fs_blob() { # @test
	cat >test.md <<-EOM
		newest body
	EOM

	export EDITOR="bash -c 'true'"
	run_zit checkin -organize test.md </dev/null
	assert_success
	assert_output - <<-EOM
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "test"]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:organize, user_story:editor, user_story:external_ids
function checkin_explicit_organize_include_untracked_fs_blob_change_description() { # @test
	cat >test.md <<-EOM
		newest body
	EOM

	cat >desired_end_state.md <<-EOM
		  - [test.md some_tag] a different description
	EOM

	export EDITOR="bash -c 'cat desired_end_state.md >\$0'"
	run_zit checkin -organize test.md </dev/null
	assert_success
	assert_output - <<-EOM
		[some_tag @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "a different description" some_tag]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:organize, user_story:editor, user_story:external_ids
function checkin_dot_organize_include_untracked_fs_blob() { # @test
	cat >test.md <<-EOM
		newest body
	EOM

	export EDITOR="bash -c 'true'"
	run_zit checkin -organize . </dev/null
	assert_success
	assert_output_unsorted - <<-EOM
		[etikett @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-one @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[etikett-two @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[!md @77f414a7068e223113928615caf1b11edd5bd6e8312eea8cdbaff37084b1d10b !toml-type-v1]
		[one/dos @b5c4fbaac3b71657edee74de4b947f13dfa104715feb8bab7cfa4dd47cafa3db !md "dos wildly different" etikett-two]
		[one/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "wildly different" etikett-one]
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !md "test"]
		[zz-archive @b8cd0eaa1891284eafdf99d3acc2007a3d4396e8a7282335f707d99825388a93]
		[zz @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:organize, user_story:editor, user_story:external_ids
function checkin_dot_include_untracked_fs_blob_with_spaces() { # @test
	cat >"test with spaces.txt" <<-EOM
		newest body
	EOM

	run_zit checkin "test with spaces.txt" </dev/null
	assert_success
	assert_output_unsorted - <<-EOM
		[!txt @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !txt "test with spaces"]
	EOM
}

# bats test_tags=user_story:fs_blobs, user_story:organize, user_story:editor, user_story:external_ids
function checkin_dot_organize_include_untracked_fs_blob_with_spaces() { # @test
	cat >"test with spaces.txt" <<-EOM
		newest body
	EOM

	export EDITOR="bash -c 'true'"
	run_zit checkin -organize "test with spaces.txt" </dev/null
	assert_success
	assert_output_unsorted - <<-EOM
		[!txt @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 !toml-type-v1]
		[two/uno @d2b258fadce18f2de6356bead0c773ca785237cad5009925a3cf1a77603847fc !txt "test with spaces"]
	EOM
}

# bats test_tags=user_story:organize,user_story:workspace
function checkin_explicit_workspace_delete_files { # @test
	# shellcheck disable=SC2317
	function editor() (
		sed -i "s/tags = \\[]/tags = ['zz-inbox']/" "$0"
		# sed -i "/type = '!md'/a tags = 'hello'" "$0"
	)

	export -f editor

	export EDITOR="/bin/bash -c 'editor \$0'"
	run_zit edit-config
	assert_success
	assert_output - <<-EOM
		[konfig @920a6a8fe55112968d75a2c77961a311343cfd62cdcc2305aff913afee7fa638 !toml-config-v1]
	EOM

	cat >.zit-workspace <<-EOM
		---
		! toml-workspace_config-v0
		---

		query = "today"

		[defaults]
		tags = ["today"]
	EOM

	run_zit info-workspace query
	assert_success
	assert_output 'today'

	echo "file one" >1.md
	echo "file two" >2.md

	function editor() {
		# shellcheck disable=SC2317
		cat - >"$1" <<-EOM
			---
			% instructions: to prevent an object from being checked in, delete it entirely
			% delete:true delete once checked in
			- today
			---

			- [1.md]
			- [2.md]
		EOM
	}

	export -f editor

	# shellcheck disable=SC2016
	export EDITOR='bash -c "editor $0"'

	run_zit checkin -organize -delete 1.md 2.md
	assert_success
	assert_output - <<-EOM
		[today @e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
		[two/uno @198cef2c92e80b728ae28c9978e64381fa18d9b31adf2068ca63b1d53153cf95 !md "1" today]
		[one/tres @7c78e911130103a9d7760788394a4467e20bf854810f915a99b9c244b266717e !md "2" today]
		          deleted [1.md]
		          deleted [2.md]
	EOM
}
