package rpc

import (
	"regexp"
	"fmt"
	//"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"
	"galaxyotc/common/errors"
)

//func NewError(err error) error {
//	if err == nil {
//		return err
//	}
//
//	srvErr := errors.NewCode(errors.ErrInternal, 4).WithCause(err)
//	e, ok := err.(*errors.Error)
//	if !ok {
//		log.Warn(srvErr)
//		return grpc.Errorf(codes.Code(errors.ErrInternal), err.Error())
//	}
//	switch {
//	case e.Code < 1024:
//		log.Warn(err)
//	default:
//		log.Info(err)
//	}
//	return grpc.Errorf(codes.Code(e.Code), err.Error())
//}

var rpcErrorRegexp = regexp.MustCompile(`CLIENT error: (.*)`)

func ParseError(err error) *errors.Error {
	if e, ok := err.(*errors.Error); ok {
		return e
	}
	fmt.Println(err.Error())
	match := rpcErrorRegexp.FindStringSubmatch(err.Error())
	if len(match) != 2 {
		return errors.NewCode(errors.ErrInternal, 4).WithCause(err)
	}
	desc := match[1]
	return errors.New(desc)
}
