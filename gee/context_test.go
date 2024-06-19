package gee

import "testing"

func TestNext(t *testing.T) {
	context := &Context{
		handlers: []HandlerFunc{A, B},
		index:    -1, // Start before the first handler
	}
	context.Next()

}
func A(c *Context) {
	println("part1")
	c.Next()
	println("part2")
}
func B(c *Context) {
	println("part3")
	c.Next()
	println("part4")
}
