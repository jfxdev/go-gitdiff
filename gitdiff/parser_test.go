	"io"
	t.Run("read", func(t *testing.T) {
		if err := p.Next(); err != nil {
			t.Fatalf("error advancing parser: %v", err)

		line := p.Line(0)
		if err := p.Next(); err != nil {
			t.Fatalf("error advancing parser: %v", err)

		line = p.Line(0)
		if err := p.Next(); err != nil {
			t.Fatalf("error advancing parser: %v", err)
		}
		line = p.Line(0)
		if line != "the third line\n" {
			t.Fatalf("incorrect third line: %s", line)

		if err := p.Next(); err != io.EOF {
			t.Fatalf("expected EOF, but got: %v", err)
	})

	t.Run("peek", func(t *testing.T) {
		p := newParser()
		if err := p.Next(); err != nil {
			t.Fatalf("error advancing parser: %v", err)

		line := p.Line(1)
		if line != "the second line\n" {
		if err := p.Next(); err != nil {
			t.Fatalf("error advancing parser: %v", err)

		line = p.Line(0)
		if line != "the second line\n" {
		"trailingComment": {
			Input: "@@ -21,5 +28,9 @@ func test(n int) {\n",
				Comment:     "func test(n int) {",
			p := &parser{r: bufio.NewReader(strings.NewReader(test.Input))}
			p.Next()

			frag, err := p.ParseFragmentHeader()
			if !reflect.DeepEqual(test.Output, frag) {
				t.Fatalf("incorrect fragment\nexpected: %+v\nactual: %+v", test.Output, frag)