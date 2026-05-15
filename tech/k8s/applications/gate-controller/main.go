package main

import "k8s.local.io/networking/resource"

func main() {
	wg, ctx, cancel := resource.NewAppContext()
	defer cancel()

	envLoader, err := resource.NewViperSecretLoader(resource.SecretOpts{
		ConfigType:  "env",
		Name:        ".env",
		SearchPaths: []string{"."},
	})
	if err != nil {
		panic(err)
	}

	s := resource.NewHttpServer(
		resource.WithAddr(":8080"),
		resource.WithServiceName("gate-controller"),
		resource.WithBasePath("/controller"),
		resource.WithSecretLoader(envLoader),
	)
	s.Run(ctx, wg)

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
