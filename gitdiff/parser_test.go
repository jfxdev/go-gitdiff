	const content = "the first line\nthe second line\nthe third line\n"
func TestParseGitFileHeader(t *testing.T) {
		Output *File
		"simpleFileChange": {
			Input: `diff --git a/dir/file.txt b/dir/file.txt
index 1c23fcc..40a1b33 100644
--- a/dir/file.txt
+++ b/dir/file.txt
`,
			Output: &File{
				OldName:      "dir/file.txt",
				NewName:      "dir/file.txt",
				OldMode:      os.FileMode(0100644),
				OldOIDPrefix: "1c23fcc",
				NewOIDPrefix: "40a1b33",
			},
		},
		"fileCreation": {
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
			p := &parser{r: bufio.NewReader(strings.NewReader(test.Input))}
			header, _ := p.Line()
			var f File
			err := p.ParseGitFileHeader(&f, header)
					t.Fatalf("expected error parsing git header, got nil")
			if test.Output != nil && !reflect.DeepEqual(f, *test.Output) {
				t.Errorf("incorrect file\nexpected: %+v\nactual: %+v", *test.Output, f)