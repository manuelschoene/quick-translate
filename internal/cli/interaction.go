package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Interface for the command to interact with the user. Hides the implementation details of the terminal struct.
type interaction interface {
	io.Writer
	confirm(question confirmation) bool
}

type confirmation struct {
	question string // The question the user must decide on
	prompt   string // The few words infront of the answer keys, which are shown in brackets after it
	fallback bool   // Fallback answer if enter is pressed directly (true for yes, false for no)
}

type terminal struct {
	in        *bufio.Reader // Buffered reader of the input stream to capture answers, which have not been asked for yet but are already provided. Not replaced per question.
	out       io.Writer     // Writer to the output stream to display messages, discarding them when silent
	prompt    io.Writer     // Writer to the output stream for questions, which are shown even when silent
	assumeYes bool          // Answers every question with yes without asking, set by -yes
}

const (
	answerYes = 'y'
	answerNo  = 'n'
)

// Creates the interaction over the given streams. Silent discards the messages, but still asks the questions.
func newTerminal(in io.Reader, out io.Writer, level verbosity, assumeYes bool) *terminal {
	messages := out
	if level == silent {
		messages = io.Discard
	}

	return &terminal{in: bufio.NewReader(in), out: messages, prompt: out, assumeYes: assumeYes}
}

// Renders the answer keys by showing the fallback answer in upper case. Does include the brackets, but no spacing around them.
func (c confirmation) answerKeys() string {
	if c.fallback {
		return fmt.Sprintf("[%c/%c]", unicode.ToUpper(answerYes), answerNo)
	}

	return fmt.Sprintf("[%c/%c]", answerYes, unicode.ToUpper(answerNo))
}

// Writes to the output stream of the terminal, allowing commands to display messages. Returns an error if the write fails.
func (t *terminal) Write(p []byte) (int, error) {
	return t.out.Write(p)
}

// Asks the user the question and reads the answer, asking again if the answer cannot be interpreted as yes or no. Returns true for yes, false for no, and the fallback answer if no answer was given at all. With -yes, nothing is asked and true is returned.
func (t *terminal) confirm(question confirmation) bool {
	if t.assumeYes {
		return true
	}

	fmt.Fprintf(t.prompt, "\n%s\n", question.question)

	for {
		fmt.Fprintf(t.prompt, "\n%s %s: ", question.prompt, question.answerKeys())

		answer, err := t.in.ReadString('\n')
		if err != nil && len(answer) == 0 {
			fmt.Fprintln(t.prompt)

			return question.fallback
		}

		switch letter(answer) {
		case answerYes:
			return true
		case answerNo:
			return false
		case 0:
			return question.fallback
		}
	}
}

// Returns the lower cased letter an answer amounts to. If the answer is longer than one letter, the number of runes it contains is returned instead, which is never zero as an empty answer must be  handled before calling this function.
func letter(answer string) rune {
	trimmed := strings.ToLower(strings.TrimSpace(answer))

	runes := []rune(trimmed)
	if len(runes) != 1 {
		return rune(len(runes))
	}

	return runes[0]
}
