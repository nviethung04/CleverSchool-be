package base

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/i18n"
	"be-lms/redis"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const SchoolRoleId = 4

type BaseRepository[T any] struct {
	queryBuilder    *QueryBuilder[T]
	mediaSync       *MediaSync[T]
	ctx             *gin.Context
	enableRedis     bool
	redisTime       time.Duration
	beforeQueryHook BeforeQueryHook // thêm trường này
	// Meili
	meiliIndex      string
	meiliFilterKeys []string
}

func NewBaseRepository[T any]() *BaseRepository[T] {
	return &BaseRepository[T]{
		queryBuilder: NewQueryBuilder[T](),
		mediaSync:    NewMediaSync[T](),
		enableRedis:  false,
		redisTime:    time.Hour * 24 * 1,
	}
}

// ====== Query Configuration Methods ======

func (r *BaseRepository[T]) SetFilter(filter map[string]interface{}) {
	r.queryBuilder.SetFilter(filter)
}

func (r *BaseRepository[T]) SetOmit(omit []string) {
	r.queryBuilder.SetOmit(omit)
}

func (r *BaseRepository[T]) SetPreload(preload []string) {
	r.queryBuilder.SetPreload(preload)
}

func (r *BaseRepository[T]) SetSort(sort map[string]string) {
	r.queryBuilder.SetSort(sort)
}

func (r *BaseRepository[T]) SetLimit(limit int) {
	r.queryBuilder.SetLimit(limit)
}

func (r *BaseRepository[T]) SetPage(page int) {
	r.queryBuilder.SetPage(page)
}

func (r *BaseRepository[T]) SetSearch(value string, fields []string) {
	r.queryBuilder.SetSearch(value, fields)
}

func (r *BaseRepository[T]) SetAlias(alias map[string]string) {
	r.queryBuilder.SetAlias(alias)
}

// ====== Cache Configuration Methods ======

func (r *BaseRepository[T]) SetEnableRedis(enable bool) {
	r.enableRedis = enable
}

func (r *BaseRepository[T]) SetRedisTime(duration time.Duration) {
	r.redisTime = duration
}

func (r *BaseRepository[T]) SetContext(c *gin.Context) {
	r.ctx = c
}

func (r *BaseRepository[T]) ApplySearch(query *gorm.DB) *gorm.DB {
	return r.queryBuilder.ApplySearch(query)
}

func (r *BaseRepository[T]) SetBeforeQueryHook(hook BeforeQueryHook) {
	r.beforeQueryHook = hook
}

func (r *BaseRepository[T]) SetMeili(index string, filterKeys []string) {
	r.meiliIndex = index
	r.meiliFilterKeys = filterKeys
}

// ====== Core CRUD Operations ======

func (r *BaseRepository[T]) GetAll() ([]T, error) {
	var entities []T

	query := db.ReplicaDB.Model(new(T))
	if r.beforeQueryHook != nil {
		query = r.beforeQueryHook.BeforeQuery(query, r.ctx)
	}
	query = r.queryBuilder.ApplyFilters(query)

	for _, relation := range r.queryBuilder.preload {
		var relationName, relationCondition string
		parts := strings.SplitN(relation, ":", 2)
		relationName = strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			relationCondition = strings.TrimSpace(parts[1])
		}

		if relationCondition != "" {
			query = query.Preload(relationName, relationCondition)
		} else {
			query = query.Preload(relationName)
		}
	}

	query = r.queryBuilder.ApplySort(query)

	err := query.Model(new(T)).Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return entities, err
}

func (r *BaseRepository[T]) FindAll() ([]T, int64, error) {
	var entities []T
	var totalCount int64

	entities, totalCount, _ = r.BeforeFindAll()

	if len(entities) > 0 {
		return entities, totalCount, nil
	}

	// Try Meili if configured and have search keyword
	if r.queryBuilder.searchValue != "" && r.meiliIndex != "" && config.LoadConfig().MeiliEnabled {
		ids, _, err := r.SearchMeiliForIDs()
		r.queryBuilder.isMeili = true

		if err == nil && len(ids) > 0 {
			newIDs := ids

			if existingVal, ok := r.queryBuilder.filter["id"]; ok {
				if existingStr, ok := existingVal.(string); ok && strings.HasPrefix(existingStr, "in:") {
					oldIDs := strings.Split(strings.TrimPrefix(existingStr, "in:"), ",")
					// Tạo map từ newIDs để check nhanh
					newSet := make(map[string]bool, len(newIDs))
					for _, id := range newIDs {
						newSet[id] = true
					}
					// Lọc lấy phần giao nhau theo thứ tự cũ
					intersection := make([]string, 0)
					for _, id := range oldIDs {
						if newSet[id] {
							intersection = append(intersection, id)
						}
					}
					newIDs = intersection
				}
			}

			if len(newIDs) > 0 {
				r.queryBuilder.filter["id"] = "in:" + strings.Join(newIDs, ",")
				orderCases := make([]string, 0, len(newIDs))
				for idx, idStr := range newIDs {
					orderCases = append(orderCases, fmt.Sprintf("WHEN id = %s THEN %d", idStr, idx))
				}

				// Todo sort by meili
				// r.SetBeforeQueryHook(orderByCaseHook{orderExpr: fmt.Sprintf("CASE %s END", strings.Join(orderCases, " "))})
			} else {
				r.queryBuilder.filter["id"] = "in:-1"
			}
		} else {
			r.queryBuilder.filter["id"] = "in:-1"
		}
	}

	query := db.ReplicaDB.Model(new(T))
	if r.beforeQueryHook != nil {
		query = r.beforeQueryHook.BeforeQuery(query, r.ctx)
	}
	query = r.queryBuilder.ApplyFilters(query)

	// Use DISTINCT to avoid counting duplicate rows from JOINs
	useMeiliOrder := r.queryBuilder.searchValue != "" && r.meiliIndex != "" && config.LoadConfig().MeiliEnabled
	if !useMeiliOrder {
		query = query.Distinct()
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query = db.ReplicaDB.Model(new(T))
	if r.beforeQueryHook != nil {
		query = r.beforeQueryHook.BeforeQuery(query, r.ctx)
	}
	query = r.queryBuilder.ApplyFiltersAndPagination(query)

	query = r.queryBuilder.ApplySort(query)
	query = r.queryBuilder.ApplyAlias(query)

	for _, relation := range r.queryBuilder.preload {
		var relationName, relationCondition string
		parts := strings.SplitN(relation, ":", 2)
		relationName = strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			relationCondition = strings.TrimSpace(parts[1])
		}

		if relationCondition != "" {
			query = query.Preload(relationName, relationCondition)
		} else {
			query = query.Preload(relationName)
		}
	}

	// Todo sort by meili
	// Avoid DISTINCT when ordering by CASE (Meili ranking), otherwise Postgres complains
	useMeiliOrder = r.queryBuilder.searchValue != "" && r.meiliIndex != "" && config.LoadConfig().MeiliEnabled
	if !useMeiliOrder {
		query = query.Distinct()
	}

	err := query.Find(&entities).Error

	if err != nil {
		return nil, 0, err
	}

	r.AfterFindAll(entities, totalCount)

	return entities, totalCount, err
}

type orderByCaseHook struct{ orderExpr string }

func (h orderByCaseHook) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	if h.orderExpr != "" {
		return query.Order(h.orderExpr)
	}
	return query
}

func (r *BaseRepository[T]) SearchMeiliForIDs() ([]string, int64, error) {
	cfg := config.LoadConfig()
	host := strings.TrimSpace(cfg.MeiliHost)
	if host == "" {
		return nil, 0, fmt.Errorf("MEILI_HOST empty")
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}

	// Build filters from current r.queryBuilder.filter using configured keys
	filters := make([]string, 0)
	for _, key := range r.meiliFilterKeys {
		if v, ok := r.queryBuilder.filter[key]; ok {
			if key == "id" {
				if str, ok2 := v.(string); ok2 && strings.HasPrefix(str, "in:") {
					ids := strings.TrimPrefix(str, "in:")
					filters = append(filters, fmt.Sprintf("id IN [%s]", ids))
				}
				continue
			}
			// quote string values
			if sv, ok := v.(string); ok {
				filters = append(filters, fmt.Sprintf("%s = '%s'", key, sv))
			} else {
				filters = append(filters, fmt.Sprintf("%s = %v", key, v))
			}
		}
	}

	type meiliReq struct {
		Q                    string   `json:"q"`
		Limit                int      `json:"limit"`
		Offset               int      `json:"offset"`
		Filter               []string `json:"filter,omitempty"`
		AttributesToRetrieve []string `json:"attributesToRetrieve,omitempty"`
	}
	type meiliRes struct {
		Hits               []map[string]interface{} `json:"hits"`
		EstimatedTotalHits *int64                   `json:"estimatedTotalHits,omitempty"`
		TotalHits          *int64                   `json:"totalHits,omitempty"`
	}

	offset := (r.queryBuilder.page - 1) * r.queryBuilder.limit
	reqBody := meiliReq{Q: r.queryBuilder.searchValue, Limit: r.queryBuilder.limit, Offset: offset, Filter: filters, AttributesToRetrieve: []string{"id"}}

	body, _ := json.Marshal(reqBody)
	url := strings.TrimRight(host, "/") + "/indexes/" + r.meiliIndex + "/search"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cfg.MeiliAPIKey != "" {
		req.Header.Set("X-Meili-API-Key", cfg.MeiliAPIKey)
		req.Header.Set("Authorization", "Bearer "+cfg.MeiliAPIKey)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf("meili http %d", resp.StatusCode)
	}
	var res meiliRes
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, 0, err
	}
	ids := make([]string, 0, len(res.Hits))
	for _, hit := range res.Hits {
		if idVal, ok := hit["id"]; ok {
			switch v := idVal.(type) {
			case float64:
				ids = append(ids, fmt.Sprintf("%d", int64(v)))
			case int64:
				ids = append(ids, fmt.Sprintf("%d", v))
			case int:
				ids = append(ids, fmt.Sprintf("%d", v))
			case string:
				ids = append(ids, v)
			}
		}
	}
	var total int64
	if res.EstimatedTotalHits != nil {
		total = *res.EstimatedTotalHits
	} else if res.TotalHits != nil {
		total = *res.TotalHits
	}
	return ids, total, nil
}

func (r *BaseRepository[T]) FindByID(id int) (*T, error) {
	var entity T

	cachedEntity, _ := r.BeforeFindById(id)

	if cachedEntity != nil {
		return cachedEntity, nil
	}

	query := db.ReplicaDB.Model(new(T))
	if r.beforeQueryHook != nil {
		query = r.beforeQueryHook.BeforeQuery(query, r.ctx)
	}
	for _, relation := range r.queryBuilder.preload {
		var relationName, relationCondition string
		parts := strings.SplitN(relation, ":", 2)
		relationName = strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			relationCondition = strings.TrimSpace(parts[1])
		}

		if relationCondition != "" {
			query = query.Preload(relationName, relationCondition)
		} else {
			query = query.Preload(relationName)
		}
	}

	query = r.queryBuilder.ApplyAlias(query)

	err := query.First(&entity, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			msg := i18n.Localize("messages.no_records_found_by_id", map[string]interface{}{"ID": id})
			return nil, fmt.Errorf(msg)
		}
		return nil, err
	}

	r.AfterFindById(id, &entity)

	return &entity, nil
}

func (r *BaseRepository[T]) FindNewByID(id int) (*T, error) {
	var entity T
	query := db.MasterDB.Model(new(T))
	if r.beforeQueryHook != nil {
		query = r.beforeQueryHook.BeforeQuery(query, r.ctx)
	}
	query = r.queryBuilder.ApplyAlias(query)

	for _, relation := range r.queryBuilder.preload {
		var relationName, relationCondition string
		parts := strings.SplitN(relation, ":", 2)
		relationName = strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			relationCondition = strings.TrimSpace(parts[1])
		}

		if relationCondition != "" {
			query = query.Preload(relationName, relationCondition)
		} else {
			query = query.Preload(relationName)
		}
	}

	err := query.First(&entity, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			msg := i18n.Localize("messages.no_records_found_by_id", map[string]interface{}{"ID": id})
			return nil, fmt.Errorf(msg)
		}
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Create(entity *T) error {
	r.BeforeCreate(entity)

	return db.MasterDB.Omit("author_id").Create(entity).Error
}

func (r *BaseRepository[T]) Update(entity *T) error {
	r.BeforeUpdate(entity)

	omit := append([]string{"created_at", "created_by", "author_id"}, r.queryBuilder.GetOmit()...)

	return db.MasterDB.Omit(omit...).Save(entity).Error
}

func (r *BaseRepository[T]) Delete(id int) error {
	var model T

	if err := db.MasterDB.First(&model, id).Error; err != nil {
		return err
	}

	if err := r.BeforeDelete(id, &model); err != nil {
		return err
	}

	if err := db.MasterDB.Omit("author_id").Save(&model).Error; err != nil {
		return err
	}

	modelVal := reflect.ValueOf(&model).Elem()
	field := modelVal.FieldByName("ID")
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("field ID not found or not settable in %T", model)
	}
	field.SetInt(int64(id))

	return db.MasterDB.Delete(&model).Error
}

func (r *BaseRepository[T]) Restore(id int) (*T, error) {
	var model T

	if err := db.MasterDB.Unscoped().First(&model, id).Error; err != nil {
		return nil, err
	}

	if err := r.BeforeRestore(id, &model); err != nil {
		return nil, err
	}

	reflectValue := reflect.ValueOf(&model).Elem()
	deletedAtField := reflectValue.FieldByName("DeletedAt")
	if deletedAtField.IsValid() {
		deletedAtField.Set(reflect.Zero(deletedAtField.Type()))
	}

	if err := db.MasterDB.Unscoped().Save(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

// ====== Cleanup Operations ======

func (r *BaseRepository[T]) PermanentlyDeleteOldRecords(before time.Time) error {
	var model T
	return db.MasterDB.
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Delete(&model).Error
}

// ====== Media Sync Operations ======

func (r *BaseRepository[T]) SyncMediaInfo() error {
	return r.mediaSync.SyncMediaInfo()
}

func (r *BaseRepository[T]) SyncMediaInfoForRecentlyUpdatedMedia(within time.Duration) error {
	return r.mediaSync.SyncMediaInfoForRecentlyUpdatedMedia(within)
}

func (r *BaseRepository[T]) SyncMediaInfoByID(id int64, infoFieldName string, dbColumnName string) error {
	return r.mediaSync.SyncMediaInfoByID(id, infoFieldName, dbColumnName)
}

// ====== Hook Methods ======

func (r *BaseRepository[T]) BeforeDelete(id int, entity *T) error {
	// Clear cache
	if r.enableRedis {
		err := redis.DeleteCacheById[T](id)
		if err != nil {
			return fmt.Errorf("failed to delete cache: %w", err)
		}
	}

	reflectValue := reflect.ValueOf(entity).Elem()
	deletedByField := reflectValue.FieldByName("DeletedBy")

	if deletedByField.IsValid() && deletedByField.CanSet() && deletedByField.Kind() == reflect.Int64 {
		var userID int64 = 0
		if userIDVal, ok := r.ctx.Get("userID"); ok {
			switch v := userIDVal.(type) {
			case int:
				userID = int64(v)
			case int64:
				userID = v
			case float64:
				userID = int64(v)
			}
		}
		deletedByField.SetInt(userID)
	}

	return nil
}

func (r *BaseRepository[T]) BeforeRestore(id int, entity *T) error {
	// Clear cache
	if r.enableRedis {
		err := redis.DeleteCacheById[T](id)
		if err != nil {
			return fmt.Errorf("failed to delete cache: %w", err)
		}
	}

	reflectValue := reflect.ValueOf(entity).Elem()
	deletedByField := reflectValue.FieldByName("DeletedBy")
	updatedByField := reflectValue.FieldByName("UpdatedBy")

	if deletedByField.IsValid() && deletedByField.CanSet() && deletedByField.Kind() == reflect.Int64 {
		deletedByField.SetInt(0)
	}

	if updatedByField.IsValid() && updatedByField.CanSet() && updatedByField.Kind() == reflect.Int64 {
		var userID int64 = 0
		if userIDVal, ok := r.ctx.Get("userID"); ok {
			switch v := userIDVal.(type) {
			case int:
				userID = int64(v)
			case int64:
				userID = v
			case float64:
				userID = int64(v)
			}
		}
		updatedByField.SetInt(userID)
	}

	return nil
}

func (r *BaseRepository[T]) BeforeFindAll() ([]T, int64, error) {
	// check and get in redis
	if r.enableRedis {
		params := map[string]interface{}{
			"filter": r.queryBuilder.filter,
			"page":   r.queryBuilder.page,
			"limit":  r.queryBuilder.limit,
		}

		if r.queryBuilder.searchValue != "" {
			params["search"] = r.queryBuilder.searchValue
		}

		for key, value := range r.queryBuilder.filter {
			params[key] = value
		}

		var cachedResult redis.CachedResult[T]
		err := redis.GetFromCacheByKey(&cachedResult, params)
		if err == nil && cachedResult.Entities != nil {
			return cachedResult.Entities, cachedResult.TotalCount, nil
		}
	}

	return []T{}, 0, nil
}

func (r *BaseRepository[T]) AfterFindAll(entities []T, totalCount int64) error {
	// check and store redis
	if r.enableRedis {
		cachedResult := redis.CachedResult[T]{
			Entities:   entities,
			TotalCount: totalCount,
		}

		params := map[string]interface{}{
			"filter": r.queryBuilder.filter,
			"page":   r.queryBuilder.page,
			"limit":  r.queryBuilder.limit,
		}

		if r.queryBuilder.searchValue != "" {
			params["search"] = r.queryBuilder.searchValue
		}

		for key, value := range r.queryBuilder.filter {
			params[key] = value
		}

		err := redis.SetToCacheByKey(cachedResult, params, r.redisTime)
		if err != nil {
			return fmt.Errorf("failed to set cache: %w", err)
		}
	}

	return nil
}

func (r *BaseRepository[T]) BeforeFindById(id int) (*T, error) {
	// check and get in redis
	if r.enableRedis {
		cachedEntity, err := redis.GetFromCacheById[*T](id)
		if err != nil {
			return nil, err
		}

		if cachedEntity != nil {
			return cachedEntity, nil
		}
	}

	return nil, nil
}

func (r *BaseRepository[T]) AfterFindById(id int, entity *T) error {
	// ✅ Store vào Redis nếu bật cache
	if r.enableRedis {
		err := redis.SetToCacheById(id, entity, r.redisTime)
		if err != nil {
			return fmt.Errorf("failed to set cache: %w", err)
		}
	}

	// ✅ Async tăng view count
	go func(id int) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Println("panic in async view update:", rec)
			}
		}()

		// Lấy kiểu T để check field
		var tmp T
		val := reflect.ValueOf(&tmp).Elem()
		field := val.FieldByName("Views")

		if field.IsValid() && field.Kind() == reflect.Int {
			// ✅ Update trực tiếp bằng ID, không dùng entity
			if err := db.MasterDB.Model(&tmp).
				Where("id = ?", id).
				UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
				fmt.Println("async update views error:", err)
			}
		}
	}(id)

	return nil
}

func (r *BaseRepository[T]) BeforeCreate(entity *T) error {
	// check and clear redis
	if r.enableRedis {
		err := redis.DeleteCacheByKey[T]()
		if err != nil {
			return err
		}
	}

	reflectValue := reflect.ValueOf(entity).Elem()
	createdByField := reflectValue.FieldByName("CreatedBy")

	// check and set created_by
	if createdByField.IsValid() && createdByField.CanSet() && createdByField.Kind() == reflect.Int64 {
		var userID int64 = 0
		if r.ctx != nil {
			if userIDVal, ok := r.ctx.Get("userID"); ok {
				switch v := userIDVal.(type) {
				case int:
					userID = int64(v)
				case int64:
					userID = v
				case float64:
					userID = int64(v)
				}
			}
		}
		createdByField.SetInt(userID)
	}

	return nil
}

func (r *BaseRepository[T]) BeforeUpdate(entity *T) error {
	// check and store redis
	if r.enableRedis {
		reflectValue := reflect.ValueOf(entity)
		reflectField := reflectValue.Elem().FieldByName("ID")

		if !reflectField.IsValid() {
			return fmt.Errorf("ID field not found in entity")
		}

		id := reflectField.Int()

		err := redis.DeleteCacheById[T](int(id))
		if err != nil {
			return fmt.Errorf("failed to delete cache: %w", err)
		}

		err = redis.DeleteCacheByKey[T]()
		if err != nil {
			return err
		}
	}

	reflectValue := reflect.ValueOf(entity).Elem()
	updatedByField := reflectValue.FieldByName("UpdatedBy")

	// check and set updated_by
	if updatedByField.IsValid() && updatedByField.CanSet() && updatedByField.Kind() == reflect.Int64 {
		var userID int64 = 0
		if r.ctx != nil {
			if userIDVal, ok := r.ctx.Get("userID"); ok {
				switch v := userIDVal.(type) {
				case int:
					userID = int64(v)
				case int64:
					userID = v
				case float64:
					userID = int64(v)
				}
			}
		}
		updatedByField.SetInt(userID)
	}

	return nil
}

type BeforeQueryHook interface {
	BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB
}

func (r *BaseRepository[T]) GetAdminSchoolId(ctx *gin.Context) int {
	if ctx == nil {
		return 0
	}

	var roleIDs []int

	if roleIDsVal, exists := ctx.Get("roleIDs"); exists {
		if ids, ok := roleIDsVal.([]int); ok {
			roleIDs = ids
		}
	}

	for _, roleID := range roleIDs {
		if roleID == SchoolRoleId {
			var schoolId int
			idVal, ok := ctx.Get("schoolID")
			if ok {
				switch v := idVal.(type) {
				case int64:
					schoolId = int(v)
				case int:
					schoolId = v
				case float64:
					schoolId = int(v)
				default:
					schoolId = 0
				}
			}

			return schoolId
		}
	}

	return 0
}
