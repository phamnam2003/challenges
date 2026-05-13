package main

import "k8s.local.io/networking/resource"

func main() {
	wg, ctx, cancel := resource.NewAppContext()
	defer cancel()

	s := resource.NewHttpServer("gate-controller")
	s.Run(ctx, wg, ":8080")

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
