Output demo/demo.gif

# https://github.com/charmbracelet/vhs/blob/main/THEMES.md
Set Theme "Catppuccin Frappe"

Set FontSize 14
Set Width 1200
Set Height 600
Set Padding 0
Set Margin 1

# --- START ---
Hide
Type "go run . --import ~/Prv/bike/history"
Enter
Sleep 0.2
Show

Sleep 1
Down@200ms 12
Up@200ms 12

Sleep 1
Type@500ms "/"
Type@500ms "/06."
Enter
Sleep 1
Escape
Sleep 1
