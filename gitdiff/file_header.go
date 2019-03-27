	"io"
	"time"
const (
	devNull = "/dev/null"
)

func (p *parser) ParseGitFileHeader() (*File, error) {
	const prefix = "diff --git "

	if !strings.HasPrefix(p.Line(0), prefix) {
		return nil, nil
	}
	header := p.Line(0)[len(prefix):]

	defaultName, err := parseGitHeaderName(header)
	if err != nil {
		return nil, p.Errorf(0, "git file header: %v", err)
	}

	f := &File{}
	for {
		end, err := parseGitHeaderData(f, p.Line(1), defaultName)
		if err != nil {
			return nil, p.Errorf(1, "git file header: %v", err)
		}
		if end {
			break
		}

		if err := p.Next(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	if f.OldName == "" && f.NewName == "" {
		if defaultName == "" {
			return nil, p.Errorf(0, "git file header: missing filename information")
		}
		f.OldName = defaultName
		f.NewName = defaultName
	}

	if (f.NewName == "" && !f.IsDelete) || (f.OldName == "" && !f.IsNew) {
		return nil, p.Errorf(0, "git file header: missing filename information")
	}

	return f, nil
}

func (p *parser) ParseTraditionalFileHeader() (*File, error) {
	const shortestValidFragHeader = "@@ -0,0 +1 @@\n"
	const (
		oldPrefix = "--- "
		newPrefix = "+++ "
	)

	oldLine, newLine := p.Line(0), p.Line(1)

	if !strings.HasPrefix(oldLine, oldPrefix) || !strings.HasPrefix(newLine, newPrefix) {
		return nil, nil
	}
	// heuristic: only a file header if followed by a (probable) fragment header
	if len(p.Line(2)) < len(shortestValidFragHeader) || !strings.HasPrefix(p.Line(2), "@@ -") {
		return nil, nil
	}

	// advance past the first line so parser is at end of header
	// no EOF check needed because we know there are >=3 valid lines
	if err := p.Next(); err != nil {
		return nil, err
	}

	oldName, _, err := parseName(oldLine[len(oldPrefix):], '\t', 0)
	if err != nil {
		return nil, p.Errorf(0, "file header: %v", err)
	}

	newName, _, err := parseName(newLine[len(newPrefix):], '\t', 0)
	if err != nil {
		return nil, p.Errorf(1, "file header: %v", err)
	}

	f := &File{}
	switch {
	case oldName == devNull || hasEpochTimestamp(oldLine):
		f.IsNew = true
		f.NewName = newName
	case newName == devNull || hasEpochTimestamp(newLine):
		f.IsDelete = true
		f.OldName = oldName
	default:
		// if old name is a prefix of new name, use that instead
		// this avoids picking variants like "file.bak" or "file~"
		if strings.HasPrefix(newName, oldName) {
			f.OldName = oldName
			f.NewName = oldName
		} else {
			f.OldName = newName
			f.NewName = newName
		}
	}

	return f, nil
}

	if len(line) > 0 && line[len(line)-1] == '\n' {
		{"@@ -", true, nil},
		{"--- ", false, parseGitHeaderOldName},
		{"+++ ", false, parseGitHeaderNewName},

// hasEpochTimestamp returns true if the string ends with a POSIX-formatted
// timestamp for the UNIX epoch after a tab character. According to git, this
// is used by GNU diff to mark creations and deletions.
func hasEpochTimestamp(s string) bool {
	const posixTimeLayout = "2006-01-02 15:04:05.9 -0700"

	start := strings.IndexRune(s, '\t')
	if start < 0 {
		return false
	}

	ts := strings.TrimSuffix(s[start+1:], "\n")

	// a valid timestamp can have optional ':' in zone specifier
	// remove that if it exists so we have a single format
	if ts[len(ts)-3] == ':' {
		ts = ts[:len(ts)-3] + ts[len(ts)-2:]
	}

	t, err := time.Parse(posixTimeLayout, ts)
	if err != nil {
		return false
	}
	if !t.Equal(time.Unix(0, 0)) {
		return false
	}
	return true
}