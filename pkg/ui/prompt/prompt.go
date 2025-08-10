package prompt

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"golang.org/x/term"
)

/*
Returns the text from the user prompt with prompt text `promptText`
and renders a help description `helpText`.
*/
func TextPrompt(promptText, helpText string) string {
	fmt.Println(helpText)
	fmt.Print(promptText)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ""
}

/*
Returns true, if the user's reply to the character prompt is "y" (case-insensitive)
and returns false otherwise.

Note that since this function sets the terminal to raw mode, all signals such as
SIGINT and SIGTERM are disabled.
*/
func YesNoPrompt(promptText, helpText string) bool {
	fmt.Println(helpText)
	fmt.Print(promptText)

	/* Switch to raw mode to read characters instantly. */
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}

	/* Display pressed character. */
	fmt.Printf("%c\n\n\r", char)

	return char == 'y' || char == 'Y'
}
