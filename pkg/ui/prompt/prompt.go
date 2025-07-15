package prompt

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"sheeper.com/fancaps-scraper-go/pkg/seq"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Style for help tips. */
var HelpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

/* Rendered help text for search query prompt. */
var SearchHelpPrompt = strings.Join([]string{
	HelpStyle.Render("Type the name of a movie, TV series, or anime you'd like to search for."),
	HelpStyle.Render(`(e.g., "Predator", "Family Guy", "Hunter x Hunter", etc.)`),
	HelpStyle.Render("Tip: You can enter just part of a title to search."),
}, "\n")

/* Returns the rendered text for the episode selection of title `title`. */
func getSelectEpisodeHelp(title *types.Title) string {
	max := strconv.Itoa(getLastEpisodeNumber(title.Episodes))

	return strings.Join([]string{
		HelpStyle.Render("Provide a range of episodes you'd like to scrape from " + "\"" + title.Name + "\""),
		HelpStyle.Render("(e.g., 1-10, 1-3, 1-, " + "-" + max + ",  etc.)"),
		HelpStyle.Render("Default: All. (1-" + max + ") [Leave empty for default]"),
		HelpStyle.Render("Tip: You can provide multiple ranges at once! (Ranges may overlap.)"),
		HelpStyle.Render("Example: \"1-5:2, 7, 6-10\" will scrape episodes 1, 3, 5, 6, 7, 8, 9, 10."),
	}, "\n")
}

/*
Returns the text from the user prompt with prompt text `promptText`
and renders a help description `helpText`.
*/
func PromptUser(promptText, helpText string) string {
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
Returns a list of titles with episodes selected by the user from titles `titles`.
If `debug` is enabled, print selected episodes and their titles.
*/
func SelectEpisodes(titles []*types.Title, debug bool) []*types.Title {
	for _, title := range titles {
		err := fmt.Errorf("selectEpisodes error placeholder: you shouldn't be here")

		/* For each (non-movie) title, prompt the user for an episode range. */
		for title.Category != types.CategoryMovie && err != nil {
			selectEpisodePrompt := "Enter Episode Range for " + title.Name + ": "
			userRange := PromptUser(selectEpisodePrompt, getSelectEpisodeHelp(title))
			if userRange == "" { // Default to all episodes if user doesn't specify a range.
				userRange = "1-" + strconv.Itoa(getLastEpisodeNumber(title.Episodes))
			}

			var episodeRange []int
			episodeRange, err = seq.ParseSequenceString(userRange, len(title.Episodes), debug)
			if err != nil {
				fmt.Fprintf(os.Stderr, "select episodes error: %v\ntry again.\n\n", err)
			} else {
				var selectedEpisodes []*types.Episode
				lastFound := 0
				for _, episodeNum := range episodeRange {
					ep, index := getEpisodeByNumber(title.Episodes, lastFound, episodeNum) // Only need to starting from the last found episode
					if ep.Name != "" && !containsEpisode(selectedEpisodes, ep) {
						selectedEpisodes = append(selectedEpisodes, ep)
						lastFound = index
					} else if containsEpisode(selectedEpisodes, ep) {
						fmt.Fprintf(os.Stderr, "select episodes warning: episode %d already selected for %s\nskipping...\n\n", episodeNum, title.Name)
					} else {
						fmt.Fprintf(os.Stderr, "select episodes error: couldn't find episode %d in %s[%d-%d]\nskipping...\n\n", episodeNum, title.Name, lastFound, len(title.Episodes))
					}
				}
				title.Episodes = selectedEpisodes
				err = nil
			}
		}
	}

	if debug {
		fmt.Println("\nSELECTED EPISODES:")
		for _, title := range titles {
			fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Link)
			for _, episode := range title.Episodes {
				fmt.Printf("\t%s -> %s\n", episode.Name, episode.Link)
			}
		}
	}

	return titles
}

/*
Returns an episode from title `title` by the episode number `episodeNum`,
starting from `start` and its index in `title`.
Returns an empty set if not found.
*/
func getEpisodeByNumber(episodes []*types.Episode, start int, episodeNum int) (*types.Episode, int) {
	re := regexp.MustCompile(fmt.Sprintf(`^Episode.*?\b%d\b.*of`, episodeNum))

	for _, ep := range episodes[start:] {
		if re.MatchString(ep.Name) {
			return ep, start
		}
		start += 1
	}

	return &types.Episode{}, -1
}

/*
Returns the last episode number of episodes `episodes`.
Returns -1, if the last episode number could not be found.
*/
func getLastEpisodeNumber(episodes []*types.Episode) int {
	if len(episodes) == 0 {
		return -1
	}

	re := regexp.MustCompile(`^Episode.*?(\d+)(?:\D+(\d+))*\s+of`)

	lastEpisodeName := episodes[len(episodes)-1].Name
	matches := re.FindStringSubmatch(lastEpisodeName)

	/* Get the last match. (i.e., the last episode number) */
	if len(matches) >= 2 {
		for i := len(matches) - 1; i >= 1; i-- {
			if matches[i] != "" {
				if num, err := strconv.Atoi(matches[i]); err == nil {
					return num
				}
			}
		}
	}

	return -1
}

/* Returns true, if `e` is in `episodes`, and returns false otherwise. */
func containsEpisode(episodes []*types.Episode, e *types.Episode) bool {
	for _, episode := range episodes {
		if episode.Link == e.Link {
			return true
		}
	}

	return false
}
