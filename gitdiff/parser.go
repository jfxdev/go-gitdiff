	"os"
	fileHeaderPrefix = "diff --git "

	devNull = "/dev/null"
	// TODO(bkeyes): parse header line for filename
	// necessary to get the filename for mode changes or add/rm empty files

	for {
		line, err := p.PeekLine()
		if err != nil {
			return err
		}

		more, err := parseGitHeaderLine(f, line)
		if err != nil {
			return p.Errorf("header: %v", err)
		}
		if !more {
			break
		}
		p.Line()
	}

	return nil

func parseGitHeaderLine(f *File, line string) (next bool, err error) {
	match := func(s string) bool {
		if strings.HasPrefix(line, s) {
			line = line[len(s):]
			return true
		}
		return false
	}

	if line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}

	switch {
	case match(fragmentHeaderPrefix):
		// start of a fragment indicates the end of the header
		return false, nil

	case match(oldFilePrefix):
		existing := f.OldName

		f.OldName, _, err = parseName(line, '\t')
		if err == nil {
			err = verifyName(f.OldName, existing, f.IsNew, "old")
		}

	case match(newFilePrefix):
		existing := f.NewName

		f.NewName, _, err = parseName(line, '\t')
		if err == nil {
			err = verifyName(f.NewName, existing, f.IsDelete, "new")
		}

	case match("old mode "):
		f.OldMode, err = parseModeLine(line)

	case match("new mode "):
		f.NewMode, err = parseModeLine(line)

	case match("deleted file mode "):
		// TODO(bkeyes): maybe set old name from default?
		f.IsDelete = true
		f.OldMode, err = parseModeLine(line)

	case match("new file mode "):
		// TODO(bkeyes): maybe set new name from default?
		f.IsNew = true
		f.NewMode, err = parseModeLine(line)

	case match("copy from "):
		f.IsCopy = true
		f.OldName, _, err = parseName(line, 0)

	case match("copy to "):
		f.IsCopy = true
		f.NewName, _, err = parseName(line, 0)

	case match("rename old "):
		f.IsRename = true
		f.OldName, _, err = parseName(line, 0)

	case match("rename new "):
		f.IsRename = true
		f.NewName, _, err = parseName(line, 0)

	case match("rename from "):
		f.IsRename = true
		f.OldName, _, err = parseName(line, 0)

	case match("rename to "):
		f.IsRename = true
		f.NewName, _, err = parseName(line, 0)

	case match("similarity index "):
		f.Score, err = parseScoreLine(line)

	case match("dissimilarity index "):
		f.Score, err = parseScoreLine(line)

	case match("index "):

	default:
		// unknown line also indicates the end of the header
		// this usually happens if the diff is empty
		return false, nil
	}

	return err == nil, err
}

func parseModeLine(s string) (os.FileMode, error) {
	mode, err := strconv.ParseInt(s, 8, 32)
	if err != nil {
		nerr := err.(*strconv.NumError)
		return os.FileMode(0), fmt.Errorf("invalid mode line: %v", nerr.Err)
	}

	return os.FileMode(mode), nil
}

func parseScoreLine(s string) (int, error) {
	score, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		nerr := err.(*strconv.NumError)
		return 0, fmt.Errorf("invalid score line: %v", nerr.Err)
	}

	if score >= 100 {
		score = 0
	}
	return int(score), nil
}

// parseName extracts a file name from the start of a string and returns the
// name and the index of the first character after the name. If the name is
// unquoted and term is non-0, parsing stops at the first occurance of term.
// Otherwise parsing of unquoted names stops at the first space or tab.
func parseName(s string, term byte) (name string, n int, err error) {
	// TODO(bkeyes): remove double forward slashes in parsed named

	if len(s) > 0 && s[0] == '"' {
		// find matching end quote and then unquote the section
		for n = 1; n < len(s); n++ {
			if s[n] == '"' && s[n-1] != '\\' {
				break
			}
		}
		if n == 1 {
			err = fmt.Errorf("missing name")
			return
		}
		n++
		name, err = strconv.Unquote(s[:n])
		return
	}

	for n = 0; n < len(s); n++ {
		if term > 0 && s[n] == term {
			break
		}
		if term == 0 && (s[n] == ' ' || s[n] == '\t') {
			break
		}
	}
	if n == 0 {
		err = fmt.Errorf("missing name")
	}
	return
}

// verifyName checks parsed names against state set by previous header lines
func verifyName(parsed, existing string, isNull bool, side string) error {
	if existing != "" {
		if isNull {
			return fmt.Errorf("expected %s, got %s", devNull, existing)
		}
		if existing != parsed {
			return fmt.Errorf("inconsistent %s filename", side)
		}
	}
	if isNull && parsed != devNull {
		return fmt.Errorf("expected %s", devNull)
	}
	return nil
}