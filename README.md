<div align="center">

# *fancaps-scraper-go!*

</div>

## Roadmap
- New CLI Flags:
  - `-t, --titles <title_url|title_name>...`
  - `-e, --episodes <episode_ranges>...`
  - `-I, -images <image_ranges>...`
  - `-i, --input <filename.txt|filename.json|filename.csv|filename.yaml>` (See Supported Filetypes)
- TUI QoL Improvements:
  - Prompt for image ranges per comma-separated value.
    - e.g., Given episode ranges: `2-4`, `7`, `8-12:2`, prompt the user to specify optional image ranges for each of the episode ranges (`2-4`, `7`, `8-12:2`). For instance, if the user specifies '1-50' for the first prompt, '100-200,250' for the second prompt, and '1-:2' for the third prompt, then the first, second, and third prompts will apply to episode ranges `2-4`, `7`, `8-12:2`, respectively.
  - Viewport.
  - Scrollbar.
  - Show item count/total.
    - e.g., Display `18/50 (36%)` on the bottom-right corner.

## IMPORTANT NOTE
**DO NOT** USE THIS PROJECT (*YET*).

It is currently <ins>unfinished</ins>.

This README will be changed to reflect the state of the project in the future.

TODO (MENTION THIS TO WINDOWS USERS): WindowSizeMsg is used to report the terminal size. Note that **Windows** does not have support for reporting when resizes occur as it does not support the SIGWINCH signal. 

*Stay tuned~*
