package quotes

import "math/rand"

type Service interface {
	RandomQuote() Quote
}

type Quote struct {
	text string
}

func (q Quote) Text() string {
	return q.text
}

type quoteRepo struct {
	quotes []Quote
}

func NewService() Service {
	var quotes = []Quote{
		{
			text: "Never decide you are smart enough. Be wise enough to recognize that there is always more to learn",
		},
		{
			text: "Intend to be as wise as nature, for she never gets pace or cadence wrong.",
		},
		{
			text: "A loving heart is the truest wisdom.",
		},
		{
			text: "The worst part of being okay is that okay is far from happy.",
		},
		{
			text: "Pain is inevitable. Suffering is optional.",
		},
		{
			text: "Wisdom is trusting the timing of the universe.",
		},
		{
			text: "Wise is the one who walks against the grain.",
		},
		{
			text: "To produce a mighty book, you must choose a mighty theme.",
		},
	}

	return &quoteRepo{
		quotes: quotes,
	}
}

func (r *quoteRepo) RandomQuote() Quote {
	randomIndex := rand.Int() % len(r.quotes)
	return r.quotes[randomIndex]
}
