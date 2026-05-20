package utils

import (
	"be-Clever School/config"
	"be-Clever School/i18n"
	"be-Clever School/prot"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

// filterProtoFields: lọc chỉ giữ các field theo ?fields=id,title
func filterProtoFields(msg proto.Message, onlyFields string) proto.Message {
	if onlyFields == "" {
		return msg
	}

	allowed := map[string]bool{}
	for _, f := range strings.Split(onlyFields, ",") {
		key := strings.TrimSpace(f)
		if key != "" {
			allowed[key] = true
		}
	}

	return filterStruct(msg, allowed)
}

func filterStruct(msg proto.Message, allowed map[string]bool) proto.Message {
	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	newMsg := reflect.New(t).Interface().(proto.Message)
	newVal := reflect.ValueOf(newMsg).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		origVal := v.Field(i)

		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = strings.ToLower(field.Name)
		}

		if origVal.Kind() == reflect.Slice && origVal.Type().Elem().Kind() == reflect.Ptr {
			if !allowed[name] {
				continue
			}
			newSlice := reflect.MakeSlice(origVal.Type(), 0, origVal.Len())
			for j := 0; j < origVal.Len(); j++ {
				elem := origVal.Index(j).Interface()
				if pm, ok := elem.(proto.Message); ok {
					filtered := filterStruct(pm, allowed)
					newSlice = reflect.Append(newSlice, reflect.ValueOf(filtered))
				}
			}
			newVal.Field(i).Set(newSlice)
			continue
		}

		if origVal.Kind() == reflect.Ptr && origVal.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
			if !allowed[name] {
				continue
			}
			if !origVal.IsNil() {
				filtered := filterStruct(origVal.Interface().(proto.Message), allowed)
				newVal.Field(i).Set(reflect.ValueOf(filtered))
			}
			continue
		}

		if allowed[name] {
			newVal.Field(i).Set(origVal)
		}
	}

	return newMsg
}

func Respond(c *gin.Context, data interface{}, err error, message string, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 && statusCode[0] != 0 {
		code = statusCode[0]
	}

	// === Trường hợp lỗi ===
	if err != nil {
		if code == 200 {
			code = http.StatusInternalServerError
		}

		if message == "" {
			message = "System error"
		}

		resp := &prot.BaseResponse{
			Code:    int32(code),
			Message: i18n.Localize(message),
			Error:   "System error",
		}

		config.Log.Info("System error...")
		errorMessage := err.Error()
		if strings.HasPrefix(errorMessage, "ERROR:") {
			reqURL := c.Request.Method + " " + c.Request.RequestURI
			config.Log.Errorf("URL: %s", reqURL)
			config.Log.Error(errorMessage)
		} else {
			config.Log.Info(errorMessage)
		}

		acceptHeader := c.GetHeader("Accept")
		if strings.Contains(acceptHeader, "application/x-protobuf") {
			protoData, err := proto.Marshal(resp)
			if err == nil {
				c.Data(code, "application/x-protobuf", protoData)
				return
			}
		}

		c.JSON(code, resp)
		return
	}

	// === Trường hợp thành công ===
	fields := c.Query("fields")

	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/x-protobuf") {
		br := &prot.BaseResponse{
			Code:    int32(code),
			Message: "success",
		}
		if msg, ok := data.(proto.Message); ok {
			if fields != "" {
				msg = filterProtoFields(msg, fields)
			}
			anyMsg, err := anypb.New(msg)
			if err == nil {
				br.Data = anyMsg
			}
		}

		out, err := proto.Marshal(br)
		if err == nil {
			c.Data(code, "application/x-protobuf", out)
			return
		}
	}

	// ✅ JSON fallback - luôn marshal đầy đủ field
	if msg, ok := data.(proto.Message); ok {
		if fields != "" {
			msg = filterProtoFields(msg, fields)
		}

		br := &prot.BaseResponse{
			Code:    int32(code),
			Message: "success",
		}

		anyMsg, err := anypb.New(msg)
		if err == nil {
			br.Data = anyMsg
		}

		var marshalOptions protojson.MarshalOptions

		if fields != "" {
			// chỉ giữ field được phép
			marshalOptions = protojson.MarshalOptions{
				EmitUnpopulated: false,
				UseProtoNames:   true,
			}
		} else {
			// Marshal tất cả field
			marshalOptions = protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   true,
			}
		}

		jsonBytes, err := marshalOptions.Marshal(br)
		if err == nil {
			c.Data(code, "application/json", jsonBytes)
			return
		}
	}

	config.Log.Warnf("Không thể format protobuf – URL: %s", c.Request.RequestURI)

	// Nếu không phải protobuf, trả JSON bình thường
	c.JSON(http.StatusOK, gin.H{
		"code":    int32(code),
		"message": "success",
		"data":    data,
	})
}
