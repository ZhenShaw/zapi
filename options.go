package zapi

type ApiOption interface {
	apply(*Api)
}

// apiOptFunc wraps a func so it satisfies the ApiOption interface.
type apiOptFunc func(*Api)

func (f apiOptFunc) apply(api *Api) {
	f(api)
}

func OptionApiName(name string) ApiOption {
	return apiOptFunc(func(api *Api) {
		api.name = name
	})
}
