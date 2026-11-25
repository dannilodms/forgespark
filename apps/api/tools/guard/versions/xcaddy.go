package versions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

// word is one word of a Dockerfile with the line it is on.
type word struct {
	text string
	line int
}

// shellOperators end the command that xcaddy runs within a `RUN`.
var shellOperators = []string{"&&", "||", ";", "|"}

// xcaddyBuilds judges what a `RUN xcaddy build` builds into the Caddy image:
// the Caddy version, when one is named, and the version of each plugin, as
// `--with module@version`. Both are Go module versions. A command that goes
// on over several lines is read whole.
func xcaddyBuilds(content string) []report.Finding {
	var findings []report.Finding
	for _, instruction := range instructionsOf(content) {
		findings = append(findings, judgeXcaddy(instruction)...)
	}
	return findings
}

// instructionsOf splits a Dockerfile into its instructions, joining the
// lines a backslash continues. A comment line is left out.
func instructionsOf(content string) [][]word {
	instructions := [][]word{nil}
	for number, line := range strings.Split(content, "\n") {
		instructions = addLine(instructions, number+1, line)
	}
	return instructions
}

// addLine adds the words of a line to the last instruction, and starts the
// next one unless the line ends with a backslash.
func addLine(instructions [][]word, number int, line string) [][]word {
	text, continued := strings.CutSuffix(strings.TrimSpace(line), `\`)
	if strings.HasPrefix(text, "#") {
		return instructions
	}
	last := len(instructions) - 1
	for _, field := range strings.Fields(text) {
		instructions[last] = append(instructions[last], word{text: field, line: number})
	}
	if continued {
		return instructions
	}
	return append(instructions, nil)
}

// xcaddyArguments returns what follows `xcaddy build` in an instruction, up
// to the end of that command, or nothing when the instruction runs no build.
func xcaddyArguments(instruction []word) []word {
	at := slices.IndexFunc(instruction, func(candidate word) bool { return candidate.text == "xcaddy" })
	if at < 0 || at+1 >= len(instruction) || instruction[at+1].text != "build" {
		return nil
	}
	arguments := instruction[at+2:]
	end := slices.IndexFunc(arguments, func(candidate word) bool { return slices.Contains(shellOperators, candidate.text) })
	if end < 0 {
		return arguments
	}
	return arguments[:end]
}

func judgeXcaddy(instruction []word) []report.Finding {
	arguments := xcaddyArguments(instruction)
	findings := judgeCaddyVersion(arguments)
	for index := range arguments {
		findings = append(findings, judgePlugin(arguments, index)...)
	}
	return findings
}

// judgeCaddyVersion judges the Caddy version xcaddy builds, the first
// argument when it is not a flag. With none, xcaddy builds the version of
// the builder image, which its `FROM` pins.
func judgeCaddyVersion(arguments []word) []report.Finding {
	if len(arguments) == 0 || strings.HasPrefix(arguments[0].text, "-") || isExactModuleVersion(arguments[0].text) {
		return nil
	}
	return []report.Finding{inexact(arguments[0].line, "the Caddy that xcaddy builds", arguments[0].text,
		"write the version of one release, as v2.11.9, or none to build the version of the builder image")}
}

// judgePlugin judges the argument at index when it is a `--with`: the
// plugin that follows it, as `module@version`, maybe with `=replacement`.
func judgePlugin(arguments []word, index int) []report.Finding {
	if arguments[index].text != "--with" || index+1 >= len(arguments) {
		return nil
	}
	plugin := arguments[index+1]
	pinned, _, _ := strings.Cut(plugin.text, "=")
	module, version, hasVersion := strings.Cut(pinned, "@")
	if !hasVersion {
		return []report.Finding{{
			Line:    plugin.line,
			Message: fmt.Sprintf("the plugin %s names no version: write it as %s@v1.2.3%s", module, module, pointer),
		}}
	}
	if isExactModuleVersion(version) {
		return nil
	}
	return []report.Finding{inexact(plugin.line, "the plugin "+module, version, directives["require"].advice)}
}

// isExactModuleVersion tells whether a Go module version names one release
// or one commit.
func isExactModuleVersion(version string) bool {
	return directives["require"].exact.MatchString(version)
}
