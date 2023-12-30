package key

import (
	"context"
)

func staticWatcher(ctx context.Context, ping chan bool) {
	for {
		select {
		case <-ping:

		case <-ctx.Done():
			return
		}
	}
}

func Static(ks *Set) (cancel func(), ping chan bool, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	ping = make(chan bool)
	go staticWatcher(ctx, ping)

	return
}
