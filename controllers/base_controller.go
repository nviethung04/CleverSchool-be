package controllers

import (
	"be-lms/i18n"
	"be-lms/prot"
	"be-lms/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

// BaseController interface định nghĩa các method cần thiết cho CRUD operations
type BaseController[T any, R any, C proto.Message] interface {
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Restore(c *gin.Context)
}

// BaseService interface định nghĩa các method service cần thiết
type BaseService[T any, R any, C proto.Message] interface {
	GetAll(c *gin.Context) ([]T, int64, error)
	GetByID(c *gin.Context, id int) (*R, error)
	Create(c *gin.Context, req C) (*T, error)
	Update(c *gin.Context, req C) (*T, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*T, error)
}

// BaseResource interface định nghĩa các method resource cần thiết
type BaseResource[T any, R any] interface {
	FormatItems(items []*T) []*R
	FormatItem(item *T) *R
}

// GenericController struct chứa logic chung cho tất cả CRUD operations
type GenericController[T any, R any, C proto.Message] struct {
	service           BaseService[T, R, C]
	resource          BaseResource[T, R]
	requestFactory    func() C
	responseFactory   func(items []*R, totalCount uint64) interface{}
	respondListHook   func(c *gin.Context, items []T, totalCount int64, err error)
	respondDetailHook func(c *gin.Context, item *T)
	usedError         func() bool
}

// NewGenericController tạo một instance mới của GenericController
func NewGenericController[T any, R any, C proto.Message](
	service BaseService[T, R, C],
	resource BaseResource[T, R],
	requestFactory func() C,
	responseFactory func(items []*R, totalCount uint64) interface{},
) *GenericController[T, R, C] {
	return &GenericController[T, R, C]{
		service:         service,
		resource:        resource,
		requestFactory:  requestFactory,
		responseFactory: responseFactory,
		usedError:       func() bool { return false },
	}
}

// WithRespondListHook đăng ký hook để tuỳ biến phản hồi danh sách
func (gc *GenericController[T, R, C]) WithRespondListHook(hook func(c *gin.Context, items []T, totalCount int64, err error)) *GenericController[T, R, C] {
	gc.respondListHook = hook
	return gc
}

// WithRespondDetailHook đăng ký hook để tuỳ biến phản hồi danh sách
func (gc *GenericController[T, R, C]) WithRespondDetailHook(hook func(*gin.Context, *T)) *GenericController[T, R, C] {
	gc.respondDetailHook = hook
	return gc
}

func (gc *GenericController[T, R, C]) WithUsedError(usedError func() bool) *GenericController[T, R, C] {
	gc.usedError = usedError
	return gc
}

// GetAll xử lý GET request để lấy tất cả items
func (gc *GenericController[T, R, C]) GetAll(c *gin.Context) {
	items, totalCount, err := gc.service.GetAll(c)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	// Ưu tiên gọi hook nếu được đăng ký
	if gc.respondListHook != nil {
		gc.respondListHook(c, items, totalCount, err)
		return
	}

	var itemPtrs []*T
	for i := range items {
		itemPtrs = append(itemPtrs, &items[i])
	}

	formattedItems := gc.resource.FormatItems(itemPtrs)
	response := gc.responseFactory(formattedItems, uint64(totalCount))

	utils.Respond(c, response, err, "")
}

// GetByID xử lý GET request để lấy item theo ID
func (gc *GenericController[T, R, C]) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := gc.service.GetByID(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data", http.StatusNotFound)
		return
	}

	utils.Respond(c, item, err, "")
}

// Create xử lý POST request để tạo item mới
func (gc *GenericController[T, R, C]) Create(c *gin.Context) {
	req, err, message := utils.GetBody[C](c, gc.requestFactory)
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	item, err := gc.service.Create(c, req)

	if err != nil {
		if gc.usedError != nil && gc.usedError() {
			utils.Respond(c, nil, err, err.Error())
		} else {
			utils.Respond(c, nil, err, "messages.create_data")
		}
		return
	}

	// Ưu tiên gọi hook nếu được đăng ký
	if gc.respondDetailHook != nil {
		gc.respondDetailHook(c, item)
		return
	}

	formattedItem := gc.resource.FormatItem(item)
	utils.Respond(c, formattedItem, err, "")
}

// Update xử lý PUT request để cập nhật item
func (gc *GenericController[T, R, C]) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.id_invalid")), "messages.id_invalid", http.StatusBadRequest)
		return
	}

	// Lấy data cũ từ database
	oldData, err := gc.service.GetByID(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data", http.StatusNotFound)
		return
	}

	// Tạo request ban đầu để so sánh
	originalReq := gc.requestFactory()

	// Đọc body một lần và tái sử dụng
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.Respond(c, nil, err, "messages.data_invalid")
		return
	}

	// Track field names từ JSON body
	jsonFields := extractFieldNamesFromJSONFromBytes(bodyBytes)

	// Lấy request mới từ body bytes đã đọc
	req, err := utils.GetBodyFromBytes[C](bodyBytes, gc.requestFactory)
	if err != nil {
		utils.Respond(c, nil, err, "messages.data_invalid")
		return
	}

	// Map data cũ vào request (chỉ những field không có trong JSON)
	mapOldDataToRequest(oldData, req, originalReq, jsonFields)

	// Set ID vào request (giả sử có field Id)
	if setter, ok := any(req).(interface{ SetId(int64) }); ok {
		setter.SetId(int64(id))
	} else {
		// Fallback: sử dụng reflection để set ID
		if reflectValue := reflect.ValueOf(req).Elem(); reflectValue.IsValid() {
			if idField := reflectValue.FieldByName("Id"); idField.IsValid() && idField.CanSet() {
				idField.SetInt(int64(id))
			}
		}
	}

	item, err := gc.service.Update(c, req)
	if err != nil {
		if gc.usedError() {
			utils.Respond(c, nil, err, err.Error())
		} else {
			utils.Respond(c, nil, err, "messages.error_update_data")
		}
		return
	}

	// Ưu tiên gọi hook nếu được đăng ký
	if gc.respondDetailHook != nil {
		gc.respondDetailHook(c, item)
		return
	}

	formattedItem := gc.resource.FormatItem(item)
	utils.Respond(c, formattedItem, err, "")
}

// Delete xử lý DELETE request để xóa item
func (gc *GenericController[T, R, C]) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := gc.service.Delete(c, id)
	utils.Respond(c, &prot.DeleteResponse{
		Id: int64(id),
	}, err, "")
}

// Restore xử lý POST request để khôi phục item đã xóa
func (gc *GenericController[T, R, C]) Restore(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "", 400)
		return
	}

	item, err := gc.service.Restore(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_restore_data")
		return
	}

	// Ưu tiên gọi hook nếu được đăng ký
	if gc.respondDetailHook != nil {
		gc.respondDetailHook(c, item)
		return
	}

	formattedItem := gc.resource.FormatItem(item)
	utils.Respond(c, formattedItem, nil, "")
}

// extractFieldNamesFromJSONFromBytes trích xuất tên các field từ JSON bytes
func extractFieldNamesFromJSONFromBytes(bodyBytes []byte) map[string]bool {
	jsonFields := make(map[string]bool)

	// Parse JSON để lấy field names
	var jsonData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
		return jsonFields
	}

	// Track field names
	for fieldName := range jsonData {
		jsonFields[fieldName] = true
	}

	return jsonFields
}

// mapOldDataToRequest map data cũ vào request (chỉ những field không có trong JSON)
func mapOldDataToRequest[R any, C any](oldData *R, req C, originalReq C, jsonFields map[string]bool) {
	// Sử dụng reflection để map
	oldVal := reflect.ValueOf(oldData)
	reqVal := reflect.ValueOf(req) // Không lấy &req nữa
	originalVal := reflect.ValueOf(originalReq)

	// Lấy element nếu là pointer
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if reqVal.Kind() == reflect.Ptr {
		reqVal = reqVal.Elem()
	}
	if originalVal.Kind() == reflect.Ptr {
		originalVal = originalVal.Elem()
	}

	// Kiểm tra kiểu dữ liệu
	if oldVal.Kind() != reflect.Struct || reqVal.Kind() != reflect.Struct || originalVal.Kind() != reflect.Struct {
		return
	}

	// Chỉ map những field không có trong JSON từ data cũ
	for i := 0; i < oldVal.NumField(); i++ {
		oldField := oldVal.Field(i)
		fieldType := oldVal.Type().Field(i)
		fieldName := fieldType.Name

		// Tìm field tương ứng trong request và original
		reqField := reqVal.FieldByName(fieldName)
		originalField := originalVal.FieldByName(fieldName)

		// Kiểm tra field có thể truy cập và không phải zero value
		oldFieldAccessible := oldField.IsValid() && oldField.CanInterface()
		reqFieldAccessible := reqField.IsValid() && reqField.CanInterface()
		originalFieldAccessible := originalField.IsValid() && originalField.CanInterface()

		if oldFieldAccessible && reqFieldAccessible && originalFieldAccessible {
			// Chỉ map khi field tồn tại, có thể set
			if reqField.CanSet() {
				// Lấy JSON tag của field để so sánh với jsonFields
				jsonTag := fieldType.Tag.Get("json")
				if jsonTag == "" {
					jsonTag = fieldName // Fallback nếu không có JSON tag
				} else {
					// Xử lý trường hợp json:"name,omitempty" -> chỉ lấy "name"
					if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
						jsonTag = jsonTag[:commaIndex]
					}
				}

				// Kiểm tra field có trong JSON không
				fieldInJSON := jsonFields[jsonTag]

				// Nếu field KHÔNG có trong JSON, map data cũ
				if !fieldInJSON {
					reqField.Set(oldField)
				}
			}
		}
	}
}
