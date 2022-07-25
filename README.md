# sideline
Backport mainline changes into a distribution package

# Note
If you rely on `compare_with`, then `git` is required at invocation

# Features
* Automatically try to generate patches against the distribution source (after all distribution patches are applied)
* Support backporting certain codebase sections
* Simple modification support (ex. search and replace)
* Ability to compare to original tag, calculate changes and include commits in changelog
* Should be able to apply certain commits to the distribution source (planned)
* Support manual conflict resolution and cache the result for future reference (planned)
