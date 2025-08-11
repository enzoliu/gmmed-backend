package lazyconn

const (
	DefaultMaxRetry uint64 = 10
)

type Options struct {
	DisableConnectionCheck bool
	MaxRetry               uint64
}

type Option func(o *Options)

func MaxRetry(maxRetry uint64) Option {
	return func(o *Options) {
		o.MaxRetry = maxRetry
	}
}

func applyOptions(options []Option) *Options {
	opts := &Options{
		DisableConnectionCheck: false,
		MaxRetry:               DefaultMaxRetry,
	}
	for _, option := range options {
		option(opts)
	}
	return opts
}
