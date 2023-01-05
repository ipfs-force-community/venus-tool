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

func Register(route *gin.RouterGroup, src interface{}, dst interface{}) {
	rvSrc := reflect.ValueOf(src)

	routeInfos := Parse(dst)
	for name, routeInfo := range routeInfos {
		fn := rvSrc.MethodByName(name)
		if !fn.IsValid() {
			log.Infof("field %s has no method", name)
			continue
		}

		route.Handle(routeInfo.Method, routeInfo.Path, Wrap(fn.Interface()))
	}
}

type RouteInfo struct {
	Name        string
	Method      string
	Path        string
	HandlerType reflect.Type
}

func Parse(dst interface{}) map[string]RouteInfo {
	rtDst := reflect.TypeOf(dst)
	rvDst := reflect.ValueOf(dst)
	if rtDst.Kind() == reflect.Ptr {
		rtDst = rtDst.Elem()
		rvDst = rvDst.Elem()
	}
	if rtDst.Kind() != reflect.Struct {
		panic("dst must be a struct or a pointer to a struct")
	}

	ret := make(map[string]RouteInfo)

	for i := 0; i < rvDst.NumField(); i++ {
		rtField := rvDst.Type().Field(i)
		name := rtField.Name
		path := rtField.Tag.Get(http.MethodGet)
		method := http.MethodGet
		if path == "" {
			path = rtField.Tag.Get(http.MethodPost)
			method = http.MethodPost
		}
		if path == "" {
			path = rtField.Tag.Get(http.MethodPut)
			method = http.MethodPut
		}
		if path == "" {
			path = rtField.Tag.Get(http.MethodDelete)
			method = http.MethodDelete
		}
		if path == "" {
			log.Infof("field %s has no route comment", name)
			continue
		}

		fn := rtField.Type
		if fn.Kind() != reflect.Func {
			log.Infof("field %s is not a function", name)
			continue
		}

		ret[name] = RouteInfo{
			Name:        name,
			Method:      method,
			Path:        path,
			HandlerType: fn,
		}
	}

	return ret
}

func Wrap(fn interface{}) gin.HandlerFunc {
	fnType := reflect.TypeOf(fn)
	fnValue := reflect.ValueOf(fn)

	ctxIdx, parmIdx, retIdx, errIdx := ProcessFunc(fnType)

	return func(ctx *gin.Context) {

		in := make([]reflect.Value, fnType.NumIn())
		if ctxIdx != -1 {
			in[ctxIdx] = reflect.ValueOf(ctx)
		}
		if parmIdx != -1 {
			pType := fnType.In(parmIdx)
			pValue := reflect.New(pType)
			pInt := pValue.Interface()

			var err error

			if ctx.Request.ContentLength > 0 {
				err = ctx.ShouldBindJSON(pInt)
			} else if ctx.Request.URL.RawQuery != "" {
				err = ctx.ShouldBindQuery(pInt)
			}

			if len(ctx.Params) > 0 {
				terr := ctx.ShouldBindUri(pInt)
				if terr != nil {
					err = terr
				}
			}

			if err != nil {
				ctx.JSON(http.StatusBadRequest, NewErrResponse(err))
				return
			}

			in[parmIdx] = pValue.Elem()
		}

		out := fnValue.Call(in)

		if errIdx != -1 {
			if !out[errIdx].IsNil() {
				ctx.JSON(http.StatusInternalServerError, NewErrResponse(out[errIdx].Interface().(error)))
				return
			}
		}
		if retIdx != -1 {
			ctx.JSON(http.StatusOK, out[retIdx].Interface())
		}
	}
}

func ProcessFunc(fnType reflect.Type) (ctxIdx, parmIdx, retIdx, errIdx int) {
	ctxIdx, parmIdx, retIdx, errIdx = -1, -1, -1, -1
	if fnType.Kind() != reflect.Func {
		panic("fn must be a function")
	}
	numIn, numOut := fnType.NumIn(), fnType.NumOut()

	// check input
	switch numIn {
	case 0:
	case 1:
		if fnType.In(0) == contextType {
			ctxIdx = 0
		} else {
			parmIdx = 0
		}
	case 2:
		if fnType.In(0) == contextType {
			ctxIdx = 0
			parmIdx = 1
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
			errIdx = 0
		} else {
			retIdx = 0
		}
	case 2:
		if fnType.Out(1) == errorType {
			errIdx = 1
			retIdx = 0
		} else {
			panic("if fn has two output, the second one must be error")
		}
	default:
		panic("fn must has at most two output")
	}

	return ctxIdx, parmIdx, retIdx, errIdx
}

type Client interface {
	Do(ctx context.Context, method, path string, in, out interface{}) error
}

func Provide(cli Client, dst interface{}) {
	// dst should be a ptr to a struct
	rtDst := reflect.TypeOf(dst)
	rvDst := reflect.ValueOf(dst)

	if rtDst.Kind() != reflect.Ptr {
		panic("dst must be a pointer to a struct")
	}
	rtDst = rtDst.Elem()
	rvDst = rvDst.Elem()

	if rtDst.Kind() != reflect.Struct {
		panic("dst must be a pointer to a struct")
	}

	routeInfo := Parse(dst)

	for name := range routeInfo {
		info := routeInfo[name]
		fnType := info.HandlerType
		ctxIdx, parmIdx, retIdx, errIdx := ProcessFunc(fnType)

		fnValue := reflect.MakeFunc(fnType, func(in []reflect.Value) (out []reflect.Value) {
			out = make([]reflect.Value, fnType.NumOut())
			ctx := context.Background()
			if ctxIdx != -1 {
				ctx = in[ctxIdx].Interface().(context.Context)
			}

			var inInt interface{}
			if parmIdx != -1 {
				inInt = in[parmIdx].Interface()
			}

			var outInt interface{}
			if retIdx != -1 {
				outInt = reflect.New(fnType.Out(retIdx)).Interface()
			}

			err := cli.Do(ctx, info.Method, info.Path, inInt, outInt)
			if errIdx != -1 {
				out[errIdx] = reflect.ValueOf(&err).Elem()
			}
			if retIdx != -1 {
				rvOut := reflect.ValueOf(outInt)
				el := rvOut.Elem()
				out[retIdx] = el
			}
			return out
		})

		field := rvDst.FieldByName(info.Name)
		field.Set(fnValue)
	}

}
