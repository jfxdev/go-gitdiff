	"bufio"
	"strings"
func TestParseGitFileHeader(t *testing.T) {
	tests := map[string]struct {
		Input  string
		Output *File
		Err    bool
	}{
		"fileContentChange": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
index 1c23fcc..40a1b33 100644
--- a/dir/file.txt
+++ b/dir/file.txt
@@ -2,3 +4,5 @@
`,
			Output: &File{
				OldName:      "dir/file.txt",
				NewName:      "dir/file.txt",
				OldMode:      os.FileMode(0100644),
				OldOIDPrefix: "1c23fcc",
				NewOIDPrefix: "40a1b33",
			},
		},
		"newFile": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
new file mode 100644
index 0000000..f5711e4
--- /dev/null
+++ b/dir/file.txt
`,
			Output: &File{
				NewName:      "dir/file.txt",
				NewMode:      os.FileMode(0100644),
				OldOIDPrefix: "0000000",
				NewOIDPrefix: "f5711e4",
				IsNew:        true,
			},
		},
		"newEmptyFile": {
			Input: `diff --git a/empty.txt b/empty.txt
new file mode 100644
index 0000000..e69de29
`,
			Output: &File{
				NewName:      "empty.txt",
				NewMode:      os.FileMode(0100644),
				OldOIDPrefix: "0000000",
				NewOIDPrefix: "e69de29",
				IsNew:        true,
			},
		},
		"deleteFile": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
deleted file mode 100644
index 44cc321..0000000
--- a/dir/file.txt
+++ /dev/null
`,
			Output: &File{
				OldName:      "dir/file.txt",
				OldMode:      os.FileMode(0100644),
				OldOIDPrefix: "44cc321",
				NewOIDPrefix: "0000000",
				IsDelete:     true,
			},
		},
		"changeMode": {
			Input: `diff --git a/file.sh b/file.sh
old mode 100644
new mode 100755
`,
			Output: &File{
				OldName: "file.sh",
				NewName: "file.sh",
				OldMode: os.FileMode(0100644),
				NewMode: os.FileMode(0100755),
			},
		},
		"rename": {
			Input: `diff --git a/foo.txt b/bar.txt
similarity index 100%
rename from foo.txt
rename to bar.txt
`,
			Output: &File{
				OldName:  "foo.txt",
				NewName:  "bar.txt",
				Score:    100,
				IsRename: true,
			},
		},
		"copy": {
			Input: `diff --git a/file.txt b/copy.txt
similarity index 100%
copy from file.txt
copy to copy.txt
`,
			Output: &File{
				OldName: "file.txt",
				NewName: "copy.txt",
				Score:   100,
				IsCopy:  true,
			},
		},
		"missingDefaultFilename": {
			Input: `diff --git a/foo.sh b/bar.sh
old mode 100644
new mode 100755
`,
			Err: true,
		},
		"missingNewFilename": {
			Input: `diff --git a/file.txt b/file.txt
index 1c23fcc..40a1b33 100644
--- a/file.txt
`,
			Err: true,
		},
		"missingOldFilename": {
			Input: `diff --git a/file.txt b/file.txt
index 1c23fcc..40a1b33 100644
+++ b/file.txt
`,
			Err: true,
		},
		"invalidHeaderLine": {
			Input: `diff --git a/file.txt b/file.txt
index deadbeef
--- a/file.txt
+++ b/file.txt
`,
			Err: true,
		},
		"notGitHeader": {
			Input: `--- file.txt
+++ file.txt
@@ -0,0 +1 @@
`,
			Output: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := &parser{r: bufio.NewReader(strings.NewReader(test.Input))}
			p.Next()

			f, err := p.ParseGitFileHeader()
			if test.Err {
				if err == nil {
					t.Fatalf("expected error parsing git file header, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error parsing git file header: %v", err)
			}

			if !reflect.DeepEqual(test.Output, f) {
				t.Errorf("incorrect file\nexpected: %+v\n  actual: %+v", test.Output, f)
			}
		})
	}
}

func TestParseTraditionalFileHeader(t *testing.T) {
	tests := map[string]struct {
		Input  string
		Output *File
		Err    bool
	}{
		"fileContentChange": {
			Input: `--- dir/file_old.txt	2019-03-21 23:00:00.0 -0700
+++ dir/file_new.txt	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName: "dir/file_new.txt",
				NewName: "dir/file_new.txt",
			},
		},
		"newFile": {
			Input: `--- /dev/null	1969-12-31 17:00:00.0 -0700
+++ dir/file.txt	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				NewName: "dir/file.txt",
				IsNew:   true,
			},
		},
		"newFileTimestamp": {
			Input: `--- dir/file.txt	1969-12-31 17:00:00.0 -0700
+++ dir/file.txt	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				NewName: "dir/file.txt",
				IsNew:   true,
			},
		},
		"deleteFile": {
			Input: `--- dir/file.txt	2019-03-21 23:30:00.0 -0700
+++ /dev/null	1969-12-31 17:00:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName:  "dir/file.txt",
				IsDelete: true,
			},
		},
		"deleteFileTimestamp": {
			Input: `--- dir/file.txt	2019-03-21 23:30:00.0 -0700
+++ dir/file.txt	1969-12-31 17:00:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName:  "dir/file.txt",
				IsDelete: true,
			},
		},
		"useShortestPrefixName": {
			Input: `--- dir/file.txt	2019-03-21 23:00:00.0 -0700
+++ dir/file.txt~	2019-03-21 23:30:00.0 -0700
@@ -0,0 +1 @@
`,
			Output: &File{
				OldName: "dir/file.txt",
				NewName: "dir/file.txt",
			},
		},
		"notTraditionalHeader": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
--- a/dir/file.txt
+++ b/dir/file.txt
`,
			Output: nil,
		},
		"noUnifiedFragment": {
			Input: `--- dir/file_old.txt	2019-03-21 23:00:00.0 -0700
+++ dir/file_new.txt	2019-03-21 23:30:00.0 -0700
context line
+added line
`,
			Output: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := &parser{r: bufio.NewReader(strings.NewReader(test.Input))}
			p.Next()

			f, err := p.ParseTraditionalFileHeader()
			if test.Err {
				if err == nil {
					t.Fatalf("expected error parsing traditional file header, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error parsing traditional file header: %v", err)
			}

			if !reflect.DeepEqual(test.Output, f) {
				t.Errorf("incorrect file\nexpected: %+v\n  actual: %+v", test.Output, f)
			}
		})
	}
}

func TestParserAdvancment(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Parse    func(p *parser) error
		NextLine string
	}{
		"ParseGitFileHeader": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
index 9540595..30e6333 100644
--- a/dir/file.txt
+++ b/dir/file.txt
@@ -1,2 +1,3 @@
context line
`,
			Parse: func(p *parser) error {
				_, err := p.ParseGitFileHeader()
				return err
			},
			NextLine: "@@ -1,2 +1,3 @@\n",
		},
		"ParseTraditionalFileHeader": {
			Input: `--- dir/file.txt
+++ dir/file.txt
@@ -1,2 +1,3 @@
context line
`,
			Parse: func(p *parser) error {
				_, err := p.ParseTraditionalFileHeader()
				return err
			},
			NextLine: "@@ -1,2 +1,3 @@\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := &parser{r: bufio.NewReader(strings.NewReader(test.Input))}
			p.Next()

			if err := test.Parse(p); err != nil {
				t.Fatalf("unexpected error while parsing: %v", err)
			}

			if err := p.Next(); err != nil {
				t.Fatalf("advancing the parser after parsing returned an error: %v", err)
			}

			if test.NextLine != p.Line(0) {
				t.Errorf("incorrect next line after parsing\nexpected: %q\nactual: %q", test.NextLine, p.Line(0))
			}
		})
	}
}


func TestHasEpochTimestamp(t *testing.T) {
	tests := map[string]struct {
		Input  string
		Output bool
	}{
		"utcTimestamp": {
			Input:  "+++ file.txt\t1970-01-01 00:00:00 +0000\n",
			Output: true,
		},
		"utcZoneWithColon": {
			Input:  "+++ file.txt\t1970-01-01 00:00:00 +00:00\n",
			Output: true,
		},
		"utcZoneWithMilliseconds": {
			Input:  "+++ file.txt\t1970-01-01 00:00:00.000000 +00:00\n",
			Output: true,
		},
		"westTimestamp": {
			Input:  "+++ file.txt\t1969-12-31 16:00:00 -0800\n",
			Output: true,
		},
		"eastTimestamp": {
			Input:  "+++ file.txt\t1970-01-01 04:00:00 +0400\n",
			Output: true,
		},
		"noTab": {
			Input:  "+++ file.txt 1970-01-01 00:00:00 +0000\n",
			Output: false,
		},
		"invalidFormat": {
			Input:  "+++ file.txt\t1970-01-01T00:00:00Z\n",
			Output: false,
		},
		"notEpoch": {
			Input:  "+++ file.txt\t2019-03-21 12:34:56.789 -0700\n",
			Output: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output := hasEpochTimestamp(test.Input)
			if output != test.Output {
				t.Errorf("incorrect output: expected %t, actual %t", test.Output, output)
			}
		})
	}
}