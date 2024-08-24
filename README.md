# tiempo

A cli to track working time to a yaml file, by default `~/tiempo.yml`.

## Install

```
go install github.com/ljurk/tiempo@latest
```

Run `tiempo --help` to see the available commands

## i3 integration

In my setup i want to start time tracking when i start my PC and take a break if the PC is locked. In i3 its quite simple:

`~/.config/i3/config`:
```
# start timetracking
exec --no-startup-id tiempo start

# start break > lock PC > end break
bindsym $mod+Shift+x exec "tiempo start --type break; i3lock -n -c 191919; tiempo end --type break"
```

I use a custom script to shutdown my PC([see](https://github.com/ljurk/dot/blob/master/bin/.bin/powermenu)), before running the shutdown command i added `tiempo end`.
