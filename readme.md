managed my git auth backend and git push config in one place. Requires git and gh.

put shell wrapper in `~/.zshrc` or bash.
```bash
# check from shell with decalre -f sw
function sw () {
	if [ "$1" = "use" ] || [ "$2" = "use" ]; then
	   if output=$(switcheroo $@ 2> /dev/null); then
		   eval $output
		   echo "switching to $(basename $output)..."
		   return
		fi
	switcheroo $@
	fi
}
```
not required but `switcheroo.yaml` can handles setting up account aliases for quick switch and even building git configs on the fly. 
Git configs are written to `~/.local/share/switcheroo/<account>`. For private account details set env vars in the shell and the tool will resolve them, you can call them whatever you want.
```yaml
# ~/.switcheroo.yaml
switcheroo:
  gitconfig:
    - alias: work
      account: "${WORK_ACCOUNT}"
      email: "${WORK_EMAIL}"
      signingkey: "${WORK_SIGNINGKEY}"
      credentials:
        - name: "https://github.com"
          helper: "!/opt/homebrew/bin/gh auth git-credential"
        - name: "https://gits.github.com"
          helper: "!/opt/homebrew/bin/gh auth git-credential"
    - alias: personal
      acount: bllacksheep
      email: "${PERSONAL_EMAIL}"
      credentials:
        - name: "https://github.com"
          helper: "!/opt/homebrew/bin/gh auth git-credential"
        - name: "https://gits.github.com"
          helper: "!/opt/homebrew/bin/gh auth git-credential"
      default: true
```
now instead of `sw use bllacksheep`, you can do:
```
sw use work
sw use personal 
```
other options
```bash
sw list            # all known profiles *current
sw current         # *current
sw use x or alias  # sets sw.Current
sw use system      # sets sw.Current to system ~/.gitconfig, essentially disables the tool
```



