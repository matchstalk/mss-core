/*
 * @Author: lwnmengjing
 * @Date: 2021/6/2 3:26 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/2 3:26 下午
 */

package server

import (
	"fmt"
	"net"

	log "github.com/matchstalk/mss-core/logger"
)

// NewListener creates a new TCP listener bound to the given address.
func NewListener(addr string) (net.Listener, error) {
	if addr == "0" {
		return nil, nil
	}
	log.Infof("listener server is starting to listen addr: %s", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		err = fmt.Errorf("error listening on %s: %w", addr, err)
		log.Error(err, "listener server failed to listen. "+
			"You may want to disable the listener server ro use another port if it is due to conflicts.")
		return nil, err
	}
	return l, nil
}
