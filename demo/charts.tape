Output demo/charts.gif

# https://github.com/charmbracelet/vhs/blob/main/THEMES.md
Set Theme "PencilLight"

Set FontSize 14
Set Width 1000
Set Height 600
Set Padding 0
Set Margin 1

# --- start ---
Hide
Type "go run . --import $IMPORT_PATH"
Enter
Sleep 6
Show

# --- jump items ---
Down@50ms 26
Sleep 100ms
Up@50ms 26
