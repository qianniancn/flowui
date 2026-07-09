package runtime

func StartCmd[Msg any](cmd func(func(Msg)), send func(Msg)) {
	if cmd != nil {
		go cmd(send)
	}
}
