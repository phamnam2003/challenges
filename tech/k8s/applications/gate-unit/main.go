package main

import "k8s.local.io/networking/resource"

func main() {
	wg, ctx, cancel := resource.NewAppContext()
	defer cancel()

	s := resource.NewHttpServer(
		resource.WithAddr(":8081"),
		resource.WithServiceName("gate-unit"),
		resource.WithBasePath("/unit"),
	)
	s.Run(ctx, wg)

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
