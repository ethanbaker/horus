package validation

// stopWords contains words that signal a multi-step command to stop
var stopWords = []string{"stop", "block", "break", "cease", "close", "cutoff", "discontinue", "terminate", "end", "kill", "desist", "quit", "cancel", "abort", "rescind", "do away with"}

// yesWords contains words that signal an affirmative
var yesWords = []string{"absolutely", "affirmative", "all right", "amen", "aye", "beyond a doubt", "by all means", "certainly", "definitely", "even so", "exactly", "fine", "gladly", "good enough", "good", "granted", "i accept", "i concur", "i guess", "if you must", "indubitably", "just so", "most assuredly", "naturally", "of course", "ok", "okay", "positively", "precisely", "right on", "righto", "sure thing", "sure", "surely", "true", "undoubtedly", "unquestionably", "very well", "whatever", "willingly", "without fail", "y", "ya", "yea", "yeah", "yep", "yes", "yessir", "yup"}

// noWords contain words that signal a negative
var noWords = []string{"no", "nay", "nix", "never", "not", "negative", "n", "not at all", "not by any means", "by no means"}
