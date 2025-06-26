#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"
}

teardown() {
	rm_from_version "$version"
}

function complete_show { # @test
	run_zit complete show
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		tag-1.*Tag
		tag-2.*Tag
		tag-3.*Tag
		tag-4.*Tag
		tag.*Tag
	EOM
}

function complete_show_all { # @test
	skip
	run_zit complete show :z,t,b,e
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		-after
		-before
		-exclude-recognized
		-exclude-untracked
		-format.*format
		-kasten.*none or Browser
		.*InventoryList
		.*InventoryList
		.*InventoryList
		.*InventoryList
		!md.*Type
		one/dos.*Zettel: !md wow ok again
		one/uno.*Zettel: !md wow the first
		tag.*Tag
		tag.1.*Tag
		tag.2.*Tag
		tag.3.*Tag
		tag.4.*Tag
	EOM
}

function complete_show_zettels { # @test
	run_zit complete show :z
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		one/dos.*Zettel: !md wow ok again
		one/uno.*Zettel: !md wow the first
	EOM
}

function complete_show_types { # @test
	run_zit complete show :t
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		md.*Type
	EOM
}

function complete_show_tags { # @test
	run_zit complete show :e
	assert_success
	assert_output_unsorted --regexp - <<-EOM
		tag-3.*Tag
		tag-4.*Tag
	EOM
}

function complete_subcmd { # @test
	run_zit complete
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		add
		cat-alfred
		cat-blob
		cat-blob-shas
		checkin
		checkin-blob
		checkin-json
		checkout
		clean
		clone
		complete.*complete a command-line
		deinit
		diff
		dormant-add
		dormant-edit
		dormant-remove
		edit
		edit-config
		exec
		export
		find-missing
		format-blob
		format-object
		format-organize
		fsck
		import
		info
		info-repo
		info-workspace
		init
		init-archive
		init-workspace
		last
		merge-tool
		new
		organize
		peek-zettel-ids
		pull
		pull-blob-store
		push
		read-blob
		reindex
		remote-add
		revert
		save
		serve
		show
		status
		test
		write-blob
	EOM
}

function complete_complete { # @test
	run_zit complete complete
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		-bash-style.*
		-in-progress.*
	EOM
}

function complete_init_workspace { # @test
	run_zit complete init-workspace
	assert_success

	# shellcheck disable=SC2016
	assert_output --regexp -- '-query.*default query for `show`'
	# shellcheck disable=SC2016
	assert_output --regexp -- '-tags.*tags added for new objects in `checkin`, `new`, `organize`'
	# shellcheck disable=SC2016
	assert_output --regexp -- '-type.*type used for new objects in `new` and `organize`'

	run_zit complete init-workspace -tags
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		tag-1.*Tag
		tag-2.*Tag
		tag-3.*Tag
		tag-4.*Tag
		tag.*Tag
	EOM

	run_zit complete init-workspace -query
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		tag-1.*Tag
		tag-2.*Tag
		tag-3.*Tag
		tag-4.*Tag
		tag.*Tag
	EOM

	run_zit complete init-workspace -type
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		!md.*Type
	EOM

	run_zit complete -in-progress="tag" init-workspace -tags tag
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		tag-1.*Tag
		tag-2.*Tag
		tag-3.*Tag
		tag-4.*Tag
		tag.*Tag
	EOM

	mkdir -p workspaces/test

	run_zit complete -in-progress="workspaces" init-workspace -tags tag workspaces
	assert_success

	# shellcheck disable=SC2016
	assert_output_unsorted --regexp -- '-query.*default query for `show`'
	# shellcheck disable=SC2016
	assert_output_unsorted --regexp -- '-tags.*tags added for new objects in `checkin`, `new`, `organize`'
	# shellcheck disable=SC2016
	assert_output_unsorted --regexp -- 'test/.*directory'
	# shellcheck disable=SC2016
	assert_output_unsorted --regexp -- '-type.*type used for new objects in `new` and `organize`'
}

function complete_checkin { # @test
	touch wow.md
	run_zit complete checkin -organize -delete
	assert_success

	# shellcheck disable=SC2016
	assert_output --regexp -- 'wow.md.*file'

	touch wow.md
	run_zit complete checkin -organize -delete --
	assert_success

	# shellcheck disable=SC2016
	assert_output --regexp -- 'wow.md.*file'
}
