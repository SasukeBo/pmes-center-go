package errormap

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/SasukeBo/log"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"html/template"
	"net/http"
)

type object map[string]interface{}
type Lang string
type langMap map[Lang]string

var (
	ZH_CN Lang = "zh_cn"
	EN    Lang = "en"

	internalErr = langMap{
		ZH_CN: "系统错误。",
		EN:    "Internal error.",
	}

	errStore = make(errorStore)
)

type errorTemplate struct {
	StatusCode int
	Languages  langMap
}

type errorStore map[string]errorTemplate

// register 注册错误代码
// errorCode 为错误代码
// statusCode 为http状态码
// languages 为 语言-模板 构成的hash表
func register(errorCode string, statusCode int, languages langMap) {
	errStore[errorCode] = errorTemplate{
		StatusCode: statusCode,
		Languages:  languages,
	}
}

// ErrorPresenter 将error处理为 gqlerror.Error
// errorCode 为错误代码
// lang 为 语言
// variables 为 模板参数值
func ErrorPresenter(errorCode string, lang Lang, originErr error, variables ...interface{}) *gqlerror.Error {
	errTemplate := errStore[errorCode]
	statusCode := errTemplate.StatusCode
	tmp := errTemplate.Languages[lang]

	tmpl, err := template.New("error").Parse(tmp)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot compile error message template %s: %v\n", tmp, err))
		log.Errorln(err)
		return &gqlerror.Error{
			Message: internalErr[lang],
			Extensions: object{
				"originErr": err.Error(),
				"code":      http.StatusInternalServerError,
			},
		}
	}

	argValues := parseArguments(variables, lang)
	buf := bytes.NewBufferString("")
	if err := tmpl.Execute(buf, argValues); err != nil {
		err = errors.New(fmt.Sprintf("Execute template failed with error: %v\n", err))
		log.Errorln(err)
		return &gqlerror.Error{
			Message: internalErr[lang],
			Extensions: object{
				"originErr": err.Error(),
				"code":      http.StatusInternalServerError,
			},
		}

	}

	errMessage := ""
	if originErr != nil {
		errMessage = originErr.Error()
	}
	return &gqlerror.Error{
		Message: buf.String(),
		Extensions: object{
			"originErr": errMessage,
			"code":      statusCode,
		},
	}
}

// parseArguments 解析参数，参数也会根据语言而返回对应值
// variables 参数数组
// lang 语言
func parseArguments(variables []interface{}, lang Lang) interface{} {
	out := make(map[string]interface{})
	for i, v := range variables {
		value := v
		if errTemplate, ok := errStore[fmt.Sprint(v)]; ok {
			if tmp, ok := errTemplate.Languages[lang]; ok {
				value = tmp
			}
		}

		out[fmt.Sprintf("field_%v", i+1)] = value
	}

	return out
}

// NewOrigin new一个origin error，并同时打印错误日志
func NewOrigin(format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)
	err := errors.New(msg)
	log.Errorln(err)
	return err
}

// SendHttpError 返回http请求错误响应信息
func SendHttpError(ctx *gin.Context, errorCode string, originErr error, variables ...interface{}) {
	var lang Lang
	langv := ctx.Request.Header.Get("Lang")
	lang = Lang(langv)
	if langv == "" {
		lang = EN
	}
	err := ErrorPresenter(errorCode, lang, originErr, variables...)
	resp := &graphql.Response{Errors: []*gqlerror.Error{err}}
	ctx.AbortWithStatusJSON(http.StatusOK, resp)
}

// SendGQLError 返回 *gqlerror.Error
func SendGQLError(ctx *gin.Context, errorCode string, originErr error, variables ...interface{}) *gqlerror.Error {
	var lang Lang
	langv := ctx.Request.Header.Get("Lang")
	lang = Lang(langv)
	if langv == "" {
		lang = EN
	}
	return ErrorPresenter(errorCode, lang, originErr, variables...)
}
