package helper

import "golang.org/x/sync/errgroup"

func Scatter(n int, fn func(i int) error) error {
	g := errgroup.Group{}

	for i := 0; i < n; i++ {
		i := i
		g.Go(func() error { return fn(i) })
	}

	return g.Wait()
}
