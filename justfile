# list commands
default:
    @just --list

# demo

alias d := demo

# build demo
demo:
    vhs demo/demo.tape
