package main

type Track struct {
	Title  string `json:"title"`
	Number string `json:"number"`
	Offset int    `json:"begin"`  // in ms
	Length int    `json:"length"` // in ms
}
