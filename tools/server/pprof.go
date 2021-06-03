/*
 * @Author: lwnmengjing
 * @Date: 2021/6/2 3:34 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/2 3:34 下午
 */

package server

import (
	"net/http"
	"net/http/pprof"
)

// PProfHandlers make pprof handlers
func PProfHandlers() map[string]http.Handler {
	return map[string]http.Handler{
		"/debug/pprof/":        http.HandlerFunc(pprof.Index),
		"/debug/pprof/cmdline": http.HandlerFunc(pprof.Cmdline),
		"/debug/pprof/profile": http.HandlerFunc(pprof.Profile),
		"/debug/pprof/symbol":  http.HandlerFunc(pprof.Symbol),
		"/debug/pprof/trace":   http.HandlerFunc(pprof.Trace),
	}
}
