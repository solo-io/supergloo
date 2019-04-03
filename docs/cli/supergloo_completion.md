---
title: "supergloo completion"
weight: 5
---
## supergloo completion

generate auto completion for your shell

### Synopsis


	Output shell completion code for the specified shell (bash or zsh).
	The shell code must be evaluated to provide interactive
	completion of supergloo commands.  This can be done by sourcing it from
	the .bash_profile.
	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2

```
supergloo completion SHELL [flags]
```

### Examples

```

	# Installing bash completion on macOS using homebrew
	## If running Bash 3.2 included with macOS
	  	brew install bash-completion
	## or, if running Bash 4.1+
	    brew install bash-completion@2
	## You may need add the completion to your completion directory
	    supergloo completion bash > $(brew --prefix)/etc/bash_completion.d/supergloo
	# Installing bash completion on Linux
	## Load the supergloo completion code for bash into the current shell
	    source <(supergloo completion bash)
	## Write bash completion code to a file and source if from .bash_profile
	    supergloo completion bash > ~/.supergloo/completion.bash.inc
	    printf "
 	     # supergloo shell completion
	      source '$HOME/.supergloo/completion.bash.inc'
	      " >> $HOME/.bash_profile
	    source $HOME/.bash_profile
	# Load the supergloo completion code for zsh[1] into the current shell
	    source <(supergloo completion zsh)
	# Set the supergloo completion code for zsh[1] to autoload on startup
	    supergloo completion zsh > "${fpath[1]}/_supergloo"
```

### Options

```
  -h, --help   help for completion
```

### SEE ALSO

* [supergloo](../supergloo)	 - CLI for Supergloo

