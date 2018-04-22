# e
The goal of e is to provide a modern line editor

E is very early in development, but when you see this line of text, it already is capable of actually editing files. This readme file was updated with e itself. Please dive into the code and send pull requests!

## How does e work?

As it is now, you can start e followed by a filename, or on its own. You will be greeted by a welcome message. On the first run and after updates it can be a good idea to check the available commands typing 'commands'.

Everything you type is saved to the command buffer, or cb for short. When an actual command is typed, it has the option to take the command buffer as its input. This way you can immediately start typing and then insert, append, replace or delete text with the i, a, r and d commands.

Use 'oops' to remove the last command buffer line, and 'oops!' to clear the whole command buffer. The print, page and search commands take one or two lines as parameters. This is likely to become nicer in the future. Don't be afraid to look in the source code for how things work and help out documenting e!
