package iotools

type errAlreadyClosed int

func (err errAlreadyClosed) Error() string { return "iotools: already closed" }
