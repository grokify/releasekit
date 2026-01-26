package run

// Git returns a Command configured for a git subcommand.
func Git(dir string, args ...string) Command {
	return Command{
		Name: "git " + firstArg(args),
		Args: append([]string{"git"}, args...),
		Dir:  dir,
	}
}

// Go returns a Command configured for a go subcommand.
func Go(dir string, args ...string) Command {
	return Command{
		Name: "go " + firstArg(args),
		Args: append([]string{"go"}, args...),
		Dir:  dir,
	}
}

// Shell returns a Command for an arbitrary shell command.
func Shell(dir string, args ...string) Command {
	return Command{
		Name: firstArg(args),
		Args: args,
		Dir:  dir,
	}
}

func firstArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
