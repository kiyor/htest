/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : logger.go

* Purpose :

* Creation Date : 03-26-2016

* Last Modified : Sat Mar 26 15:52:50 2016

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package htest

import (
	// 	"github.com/kiyor/golib"
	"github.com/op/go-logging"
)

var (
	// 	Logger = golib.NewLogger(&golib.LogOptions{
	// 		Name:      "htest",
	// 		ShowErr:   true,
	// 		ShowDebug: *Verbose,
	// 		ShowColor: true,
	// 	})
	Logger *logging.Logger
)
