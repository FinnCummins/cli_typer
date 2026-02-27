package main

// Word lists and quotes are embedded directly in the source code.
// No external files needed — the binary is fully self-contained.

import (
	"math/rand"
	"strings"
)

// Common English words (similar to monkeytype's "english" word set).
// These are deliberately simple, everyday words.
var commonWords = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "it",
	"for", "not", "on", "with", "he", "as", "you", "do", "at", "this",
	"but", "his", "by", "from", "they", "we", "say", "her", "she", "or",
	"an", "will", "my", "one", "all", "would", "there", "their", "what",
	"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
	"when", "make", "can", "like", "time", "no", "just", "him", "know",
	"take", "people", "into", "year", "your", "good", "some", "could",
	"them", "see", "other", "than", "then", "now", "look", "only", "come",
	"its", "over", "think", "also", "back", "after", "use", "two", "how",
	"our", "work", "first", "well", "way", "even", "new", "want", "because",
	"any", "these", "give", "day", "most", "us", "great", "between", "need",
	"large", "under", "never", "each", "right", "hand", "high", "place",
	"small", "found", "still", "own", "light", "word", "went", "last",
	"long", "much", "before", "turn", "move", "right", "real", "left",
	"same", "being", "world", "house", "point", "home", "old", "number",
	"start", "show", "every", "part", "find", "here", "thing", "many",
	"head", "name", "very", "through", "just", "form", "much", "line",
	"water", "been", "call", "keep", "while", "next", "program", "change",
	"room", "group", "begin", "might", "story", "along", "children", "city",
	"earth", "eye", "run", "quite", "close", "night", "open", "life",
	"walk", "white", "begin", "got", "read", "hand", "port", "spell",
	"add", "land", "must", "big", "act", "why", "ask", "men", "went",
	"air", "away", "animal", "again", "study", "help", "should", "late",
	"above", "paper", "near", "grow", "food", "learn", "plant", "cover",
	"state", "set", "try", "face", "watch", "car", "seem", "sea", "draw",
	"hard", "let", "stop", "without", "second", "city", "tree", "cross",
	"since", "hard", "pick", "fast", "several", "hold", "himself", "toward",
	"five", "step", "morning", "pass", "power", "town", "fine", "true",
	"hundred", "area", "table", "strong", "special", "mind", "behind",
	"clear", "ball", "best", "better", "dark", "rest", "early", "sort",
	"told", "money", "river", "class", "nothing", "age", "check", "game",
}

// Famous quotes for quote mode.
var quotes = []string{
	"It is a truth universally acknowledged that a single man in possession of a good fortune must be in want of a wife",
	"The only way to do great work is to love what you do",
	"In the middle of difficulty lies opportunity",
	"Not all those who wander are lost",
	"The future belongs to those who believe in the beauty of their dreams",
	"It does not do to dwell on dreams and forget to live",
	"To be yourself in a world that is constantly trying to make you something else is the greatest accomplishment",
	"In three words I can sum up everything I learned about life it goes on",
	"The greatest glory in living lies not in never falling but in rising every time we fall",
	"Life is what happens when you are busy making other plans",
	"The way to get started is to quit talking and begin doing",
	"If you look at what you have in life you will always have more",
	"If you set your goals ridiculously high and it is a failure you will fail above everyone else success",
	"You must be the change you wish to see in the world",
	"Spread love everywhere you go let no one ever come to you without leaving happier",
	"The only thing we have to fear is fear itself",
	"Darkness cannot drive out darkness only light can do that hate cannot drive out hate only love can do that",
	"Do one thing every day that scares you",
	"Well done is better than well said",
	"The best time to plant a tree was twenty years ago the second best time is now",
	"An unexamined life is not worth living",
	"Many of life great failures are people who did not realize how close they were to success when they gave up",
	"You have brains in your head you have feet in your shoes you can steer yourself any direction you choose",
	"If life were predictable it would cease to be life and be without flavor",
	"Life is a succession of lessons which must be lived to be understood",
}

// generateWords returns a slice of random words from the common word list.
// For a 60-second test we generate ~200 words (enough for even fast typists).
func generateWords(count int) []string {
	words := make([]string, count)
	for i := range words {
		words[i] = commonWords[rand.Intn(len(commonWords))]
	}
	return words
}

// getQuoteWords picks random quotes and splits them into words,
// concatenating until we have at least `minWords` words.
func getQuoteWords(minWords int) []string {
	var words []string
	for len(words) < minWords {
		quote := quotes[rand.Intn(len(quotes))]
		words = append(words, strings.Fields(quote)...)
	}
	return words
}
