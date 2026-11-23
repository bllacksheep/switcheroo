managed my git auth backend and git push config in one place. Requires git and gh
```
sw use work
sw use personal 
```
or by name `sw use bllacksheep` and view currently acitve profiles
```
sw list
sw current
```

shell wrapper
```
sw () {
	if [ "$1" = "use" ] || [ "$2" = "use" ]
	then
		eval "$(switcheroo $@ 2> /dev/null)"
		return
	fi
	switcheroo $@
}
```


