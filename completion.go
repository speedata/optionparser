package optionparser

import (
	"fmt"
	"io"
	"strings"
)

// GenerateCompletion writes a shell completion script for the given shell
// to w. Supported shells are "bash", "zsh" and "fish". programName is the
// command users invoke (for example "sp").
//
// The generated script offers tab completion for all registered long and
// short options. Options registered with a "--no-" prefix are emitted in
// both forms. Registered commands appear as positional completions.
//
// Because the parser has no knowledge of option value semantics, options
// that expect a parameter offer file completion as a best-effort default.
func (op *OptionParser) GenerateCompletion(shell, programName string, w io.Writer) error {
	if programName == "" {
		return fmt.Errorf("programName must not be empty")
	}
	switch shell {
	case "bash":
		return op.writeBashCompletion(programName, w)
	case "zsh":
		return op.writeZshCompletion(programName, w)
	case "fish":
		return op.writeFishCompletion(programName, w)
	default:
		return fmt.Errorf("unsupported shell %q (want bash, zsh or fish)", shell)
	}
}

func flagsFor(o *allowedOptions) (long []string, short []string) {
	if o.long != "" {
		long = append(long, "--"+o.long)
		if o.boolParameter {
			long = append(long, "--no-"+o.long)
		}
	}
	if o.short != "" {
		short = append(short, "-"+o.short)
	}
	return long, short
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

// sanitizeFuncName produces a string suitable as a shell function name.
func sanitizeFuncName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// ---- bash ----

func (op *OptionParser) writeBashCompletion(prog string, w io.Writer) error {
	var allFlags []string
	var valueFlags []string
	for _, o := range op.options {
		long, short := flagsFor(o)
		allFlags = append(allFlags, long...)
		allFlags = append(allFlags, short...)
		if o.param != "" {
			// only the value-taking forms; --no- toggles do not take values
			if o.long != "" {
				valueFlags = append(valueFlags, "--"+o.long)
			}
			if o.short != "" {
				valueFlags = append(valueFlags, "-"+o.short)
			}
		}
	}
	var cmds []string
	for _, c := range op.commands {
		cmds = append(cmds, c.name)
	}

	fn := "_" + sanitizeFuncName(prog) + "_completions"

	var valueCase string
	if len(valueFlags) > 0 {
		valueCase = fmt.Sprintf(`    case "${prev}" in
        %s)
            COMPREPLY=( $(compgen -f -- "${cur}") )
            return 0
            ;;
    esac
`, strings.Join(valueFlags, "|"))
	}

	_, err := fmt.Fprintf(w, `# bash completion for %s
%s() {
    local cur prev
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

%s    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "%s" -- "${cur}") )
        return 0
    fi

    COMPREPLY=( $(compgen -W "%s" -f -- "${cur}") )
    return 0
}
complete -F %s %s
`,
		prog,
		fn,
		valueCase,
		strings.Join(allFlags, " "),
		strings.Join(cmds, " "),
		fn,
		prog,
	)
	return err
}

// ---- zsh ----

// zshQuote escapes a string so it can appear inside the [description] or
// :message: section of a single-quoted _arguments spec.
func zshQuote(s string) string {
	s = oneLine(s)
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "'", `'\''`)
	s = strings.ReplaceAll(s, "[", `\[`)
	s = strings.ReplaceAll(s, "]", `\]`)
	s = strings.ReplaceAll(s, ":", `\:`)
	return s
}

func (op *OptionParser) writeZshCompletion(prog string, w io.Writer) error {
	fn := "_" + sanitizeFuncName(prog)
	var specs []string
	for _, o := range op.options {
		long, short := flagsFor(o)
		all := append([]string{}, long...)
		all = append(all, short...)
		if len(all) == 0 {
			continue
		}
		descr := zshQuote(o.helptext)
		var value string
		if o.param != "" {
			value = ":" + zshQuote(o.param) + ":_files"
		}
		if len(all) == 1 {
			specs = append(specs, fmt.Sprintf("'%s[%s]%s'", all[0], descr, value))
			continue
		}
		group := "(" + strings.Join(all, " ") + ")"
		for _, f := range all {
			// --no-foo forms never take a value, even if --foo does
			v := value
			if strings.HasPrefix(f, "--no-") {
				v = ""
			}
			specs = append(specs, fmt.Sprintf("'%s%s[%s]%s'", group, f, descr, v))
		}
	}

	var cmdLines []string
	for _, c := range op.commands {
		cmdLines = append(cmdLines, fmt.Sprintf("        '%s:%s'", c.name, zshQuote(c.helptext)))
	}

	header := "#compdef " + prog + "\n"
	body := fmt.Sprintf(`# zsh completion for %s
%s() {
    local state
    local -a commands
    commands=(
%s
    )
    _arguments -C \
        %s \
        '1: :->cmd' \
        '*::arg:->args'

    case $state in
        cmd)
            _describe -t commands '%s command' commands
            ;;
        args)
            _files
            ;;
    esac
}

if [ "${funcstack[1]-}" = "%s" ]; then
    # Loaded via fpath/autoload: run the actual completion function.
    %s "$@"
elif (( $+functions[compdef] )); then
    # Sourced directly: register with the completion system.
    compdef %s %s
fi
`,
		prog,
		fn,
		strings.Join(cmdLines, "\n"),
		strings.Join(specs, " \\\n        "),
		prog,
		fn,
		fn,
		fn,
		prog,
	)
	_, err := io.WriteString(w, header+body)
	return err
}

// ---- fish ----

func fishQuote(s string) string {
	s = oneLine(s)
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}

func (op *OptionParser) writeFishCompletion(prog string, w io.Writer) error {
	helper := "__" + sanitizeFuncName(prog) + "_needs_command"

	fmt.Fprintf(w, "# fish completion for %s\n", prog)
	fmt.Fprintf(w, "function %s\n", helper)
	fmt.Fprintf(w, "    set -l cmd (commandline -opc)\n")
	fmt.Fprintf(w, "    set -e cmd[1]\n")
	fmt.Fprintf(w, "    for token in $cmd\n")
	fmt.Fprintf(w, "        switch $token\n")
	fmt.Fprintf(w, "            case '-*'\n")
	fmt.Fprintf(w, "                continue\n")
	fmt.Fprintf(w, "            case '*'\n")
	fmt.Fprintf(w, "                return 1\n")
	fmt.Fprintf(w, "        end\n")
	fmt.Fprintf(w, "    end\n")
	fmt.Fprintf(w, "    return 0\n")
	fmt.Fprintf(w, "end\n\n")

	for _, c := range op.commands {
		fmt.Fprintf(w, "complete -c %s -n %s -f -a '%s' -d '%s'\n",
			prog, helper, c.name, fishQuote(c.helptext))
	}

	emit := func(short, long string, takesParam bool, descr string) {
		parts := []string{"complete -c " + prog}
		if short != "" {
			parts = append(parts, "-s "+short)
		}
		if long != "" {
			parts = append(parts, "-l "+long)
		}
		if takesParam {
			parts = append(parts, "-r")
		}
		if descr != "" {
			parts = append(parts, "-d '"+descr+"'")
		}
		fmt.Fprintln(w, strings.Join(parts, " "))
	}

	for _, o := range op.options {
		descr := fishQuote(o.helptext)
		emit(o.short, o.long, o.param != "", descr)
		if o.boolParameter && o.long != "" {
			emit("", "no-"+o.long, false, descr)
		}
	}
	return nil
}
