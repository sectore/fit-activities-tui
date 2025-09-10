# list commands
default:
    @just --list

alias r := run

# Runs the app by importing FIT files from given path.
[group('dev')]
run path:
    go run . --import {{ path }}

# demos

alias d := demo

# Build an animated `demo.gif`. Run this command with $IMPORT_PATH={directory-of-FIT-files} defined to point to FIT files you want to use for the demo.
[group('demo')]
demo:
    vhs demo/demo.tape

alias dc := demo-charts

[group('demo')]
demo-charts:
    vhs demo/charts.tape
