package resource

type HttpOpts struct {
	ServiceName string
	Addr        string
	BasePath    string
	SecLoader   SecretLoader
}

type HttpOptions func(*HttpOpts)

func WithServiceName(name string) HttpOptions {
	return func(opts *HttpOpts) {
		opts.ServiceName = name
	}
}

func WithAddr(addr string) HttpOptions {
	return func(opts *HttpOpts) {
		opts.Addr = addr
	}
}

func WithBasePath(basePath string) HttpOptions {
	return func(opts *HttpOpts) {
		opts.BasePath = basePath
	}
}

func WithSecretLoader(loader SecretLoader) HttpOptions {
	return func(opts *HttpOpts) {
		opts.SecLoader = loader
	}
}

// SecretOpts defines the options for configuring the secret loader.
type SecretOpts struct {
	EnvPrefix   string
	ConfigType  string
	Path        string
	Name        string
	SearchPaths []string
}
