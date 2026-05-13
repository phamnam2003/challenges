package resource

type HttpOpts struct {
	ServiceName string
	Addr        string
	BasePath    string
}

type Options func(*HttpOpts)

func WithServiceName(name string) Options {
	return func(opts *HttpOpts) {
		opts.ServiceName = name
	}
}

func WithAddr(addr string) Options {
	return func(opts *HttpOpts) {
		opts.Addr = addr
	}
}

func WithBasePath(basePath string) Options {
	return func(opts *HttpOpts) {
		opts.BasePath = basePath
	}
}
