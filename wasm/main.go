package main

import (
	"syscall/js"
)

func main() {
	wait := make(chan struct{}, 0)

	js.Global().Set("createStateQueue", js.FuncOf(createStateQueue))
	js.Global().Set("decodeStates", js.FuncOf(decodeStates))
	js.Global().Set("encodeStates", js.FuncOf(encodeStates))

	<-wait
}
