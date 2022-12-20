package route

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

var (
	errorType   = reflect.TypeOf(new(error)).Elem()
	contextType = reflect.TypeOf(new(context.Context)).Elem()
)

type ErrorResp struct {
	Err string `json:"error"`
}

func (e *ErrorResp) Error() string {
	return e.Err
}

func NewErrResponse(err error) ErrorResp {
	return ErrorResp{Err: err.Error()}
}

func Wrap(fn interface{}) gin.HandlerFunc {
	fnType := reflect.TypeOf(fn)
	fnValue := reflect.ValueOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("fn must be a function")
	}
	numIn, numOut := fnType.NumIn(), fnType.NumOut()
	hasCtx, hasParm, hasRet, hasErr := false, false, false, false

	// check input
	switch numIn {
	case 0:
	case 1:
		if fnType.In(0) == contextType {
			hasCtx = true
		}
	case 2:
		if fnType.In(0) == contextType {
			hasCtx = true
			hasParm = true
		} else {
			panic("if fn has two param, the first one must be context.Context")
		}
	default:
		panic("fn must has at most two params")
	}

	// check output
	switch numOut {
	case 0:
	case 1:
		if fnType.Out(0) == errorType {
			hasErr = true
		}
	case 2:
		if fnType.Out(1) == errorType {
			hasErr = true
			hasRet = true
		} else {
			panic("if fn has two output, the second one must be error")
		}
	default:
		panic("fn must has at most two output")
	}

	return func(ctx *gin.Context) {

		in := make([]reflect.Value, numIn)
		if hasCtx {
			in[0] = reflect.ValueOf(ctx)
		}
		if hasParm {
			pType := fnType.In(1)
			pValue := reflect.New(pType)
			pInt := pValue.Interface()

			err := ctx.ShouldBindJSON(pInt)
			if err != nil {
				err = ctx.ShouldBindUri(pInt)
			}
			if err != nil {
				err = ctx.ShouldBindQuery(pInt)
			}
			if err != nil {
				err = ctx.ShouldBind(pInt)
			}
			if err != nil {
				ctx.JSON(http.StatusBadRequest, NewErrResponse(err))
				return
			}

			in[1] = pValue.Elem()
		}

		out := fnValue.Call(in)

		if hasErr {
			if !out[numOut-1].IsNil() {
				ctx.JSON(http.StatusInternalServerError, NewErrResponse(out[numOut-1].Interface().(error)))
				return
			}
		}
		if hasRet {
			ctx.JSON(http.StatusOK, out[0].Interface())
		}
	}
}
