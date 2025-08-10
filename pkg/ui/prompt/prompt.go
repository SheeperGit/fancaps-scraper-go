package prompt

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
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
Returns true, if the user's reply to the prompt contains a "y"
as the first character (case-insensitive) and returns false otherwise.
*/
func YesNoPrompt(promptText, helpText string) bool {
	fmt.Println(helpText)
	fmt.Print(promptText)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		reply := strings.TrimSpace(scanner.Text())
		if len(reply) == 0 {
			return false
		}

		reply = strings.ToLower(string(reply[0]))
		if reply == "y" {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return false
}
