package query

import (
	"strings"
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

func TestQuery(t1 *testing.T) {
	type testCase struct {
		stackInfo                                stack_frame.Frame
		description, expected, expectedOptimized string
		defaultGenre                             ids.Genre
		inputs                                   []string
	}

	t := ui.T{T: t1}

	testCases := []testCase{
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[[test,house] home]",
			inputs:    []string{"[test, house] home"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[[test,house] home wow]",
			inputs:    []string{"[test, house] home", "wow"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[^[test,house] home wow]",
			inputs:    []string{"^[test, house] home", "wow"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[[test,house] ^home wow]",
			inputs:    []string{"[test, house] ^home", "wow"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[[test,^house] home wow]",
			inputs:    []string{"[test, ^house] home", "wow"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[[test,house] home ^wow]",
			inputs:    []string{"[test, house] home", "^wow"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[^[[test,house] home] wow]",
			inputs:    []string{"^[[test, house] home]", "wow"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "^[[test,house] home]:Zettel wow",
			inputs:    []string{"^[[test, house] home]:z", "wow"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "[!md,home]:Zettel",
			inputs:    []string{"[!md,home]:z"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "!md?Zettel",
			inputs:    []string{"!md?z"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "ducks:Tag [!md house]+?Zettel",
			inputs:    []string{"!md?z", "house+z", "ducks:e"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "ducks:Tag [!md house]?Zettel",
			inputs:    []string{"ducks:Tag [!md house]?Zettel"},
		},
		{
			stackInfo: t.MakeStackInfo(0),
			expected:  "ducks:Tag [=!md house]?Zettel",
			inputs:    []string{"ducks:Tag [=!md house]?Zettel"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: "ducks:Tag [=!md house wow]:?Zettel",
			expected:          "ducks:Tag [=!md house wow]:?Zettel",
			inputs: []string{
				"ducks:Tag [=!md house]?Zettel wow:Zettel",
			},
		},
		{ // TODO try to make this expect `one/uno.zettel`
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: "one/uno:.Zettel",
			expected:          "one/uno:.Zettel",
			inputs:            []string{"one/uno.zettel"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: "one/uno:Zettel",
			expected:          "one/uno:Zettel",
			defaultGenre:      ids.MakeGenre(genres.Zettel),
			inputs:            []string{"one/uno"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: "one/uno:Zettel",
			expected:          "one/uno:Zettel",
			inputs:            []string{"one/uno:z"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: ":Config",
			expected:          ":Config",
			inputs:            []string{":konfig"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: ":Zettel",
			expected:          ":Zettel",
			inputs:            []string{":z"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: ":Repo",
			expected:          ":Repo",
			inputs:            []string{":k"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: "one/uno:+Zettel",
			expected:          "one/uno:+Zettel",
			inputs:            []string{"one/uno+"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			expectedOptimized: "[one/dos, one/uno]:Zettel",
			expected:          "[one/dos, one/uno]:Zettel",
			inputs:            []string{"one/uno", "one/dos"},
		},
		{
			expectedOptimized: ":Type :Tag :Zettel",
			expected:          ":Type,Tag,Zettel",
			inputs:            []string{":z,t,e"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			defaultGenre:      ids.MakeGenre(genres.All()...),
			expectedOptimized: ":Blob :Type :Tag :Zettel :Config :InventoryList :Repo",
			expected:          ":Blob,Type,Tag,Zettel,Config,InventoryList,Repo",
			inputs:            []string{},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			defaultGenre:      ids.MakeGenre(genres.All()...),
			expectedOptimized: ":Blob :Type :Tag :Zettel :Config :InventoryList :Repo",
			expected:          ":Blob,Type,Tag,Zettel,Config,InventoryList,Repo",
			inputs:            []string{":"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			defaultGenre:      ids.MakeGenre(genres.All()...),
			expectedOptimized: "2109504781.792086:InventoryList",
			expected:          "2109504781.792086:InventoryList",
			inputs:            []string{"[2109504781.792086]:b"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			defaultGenre:      ids.MakeGenre(genres.All()...),
			expectedOptimized: "^etikett-two.Zettel",
			expected:          "^etikett-two.Zettel",
			inputs:            []string{"^etikett-two.z"},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			defaultGenre:      ids.MakeGenre(genres.All()...),
			expectedOptimized: "!md.Blob !md.Type !md.Tag !md.Zettel !md.Config !md.InventoryList !md.Repo",
			expected:          "!md.Blob,Type,Tag,Zettel,Config,InventoryList,Repo",
			inputs:            []string{"!md."},
		},
		{
			stackInfo:         t.MakeStackInfo(0),
			defaultGenre:      ids.MakeGenre(genres.All()...),
			expectedOptimized: "-etikett-two.Zettel",
			expected:          "-etikett-two.Zettel",
			inputs:            []string{"-etikett-two.z"},
		},

		{
			stackInfo:         t.MakeStackInfo(0),
			defaultGenre:      ids.MakeGenre(genres.All()...),
			expectedOptimized: "/repo:Repo",
			expected:          "/repo:Repo",
			inputs:            []string{"/repo:k"},
		},
	}

	for _, tc := range testCases {
		t1.Run(
			strings.Join(tc.inputs, " "),
			func(t1 *testing.T) {
				t := ui.TC{
					T:     ui.T{T: t1},
					Frame: tc.stackInfo,
				}

				sut := (&Builder{}).WithDefaultGenres(
					tc.defaultGenre,
				)

				m, err := sut.BuildQueryGroup(tc.inputs...)

				t.AssertNoError(err)
				actual := m.String()

				if tc.expected != actual {
					t.Log("expected")
					t.AssertEqual(tc.expected, actual)
				}

				if tc.expectedOptimized == "" {
					return
				}

				actualOptimized := m.StringOptimized()

				if tc.expectedOptimized != actualOptimized {
					t.Log(m.StringDebug())
					t.Log("expectedOptimized")
					t.AssertEqual(tc.expectedOptimized, actualOptimized)
				}
			},
		)
	}
}
