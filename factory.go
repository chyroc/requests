package requests

//
// type Factory[T any] struct {
// 	options []RequestOption[T]
// }
//
// func (r *Factory) New(method, url string) *Request {
// 	req := New(method, url)
// 	for _, v := range r.options {
// 		if err := v(req); err != nil {
// 			return req.SetError(err)
// 		}
// 	}
// 	return req
// }
//
// func NewFactory(options ...RequestOption) *Factory {
// 	return &Factory{options: options}
// }
