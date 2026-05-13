package main

import "k8s.local.io/networking/resource"

func main() {
	wg, ctx, cancel := resource.NewAppContext()
	defer cancel()

	s := resource.NewHttpServer(
		resource.WithAddr(":8080"),
		resource.WithServiceName("gate-controller"),
		resource.WithBasePath("/controller"),
	)
	s.Run(ctx, wg)

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
