Output demo/demo.gif

# https://github.com/charmbracelet/vhs/blob/main/THEMES.md
Set Theme "Catppuccin Frappe"

Set FontSize 14
Set Width 1000
Set Height 600
Set Padding 0
Set Margin 1

# --- start ---
Hide
Type "go run . --import $IMPORT_PATH"
Enter
Sleep 0.2
Show

# --- sort ---
Sleep 500ms
Ctrl+t
Sleep 500ms
Ctrl+d
Sleep 500ms

# --- jump items ---
Down@50ms 12
Sleep 200ms
Up@50ms 12
Sleep 200ms

# --- filter w/ open menu ---
Type@100ms "m"
Type@200ms "/"
Type@100ms "06."
Enter@100ms
Type@500ms "m"
Down@100ms 2
Up@100ms 2
Sleep 200ms
Escape
