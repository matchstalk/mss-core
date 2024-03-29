/*
 * @Author: lwnmengjing
 * @Date: 2021/6/8 2:15 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/8 2:15 下午
 */

package listener

import (
	"net/http"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	addr, certFile, keyFile string
	handler                 http.Handler
}

func setDefaultOption() options {
	return options{
		addr: ":8080",
		handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
	}
}

// WithAddr 设置addr
func WithAddr(s string) Option {
	return func(o *options) {
		o.addr = s
	}
}

// WithHandler 设置handler
func WithHandler(handler http.Handler) Option {
	return func(o *options) {
		o.handler = handler
	}
}

// WithCert 设置cert
func WithCert(s string) Option {
	return func(o *options) {
		o.certFile = s
	}
}

// WithKey 设置key
func WithKey(s string) Option {
	return func(o *options) {
		o.keyFile = s
	}
}
