
for now
export GIT_CONFIG_GLOBAL=~/.local/share/switcheroo/bllacksheep

// get names via logged in profiles
// tie configs to logged in profiles
// 
gh auth switch -u bllacksheep
gh auth refresh -h github.com
gh auth setup-git when switching
