package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/utils/client"
)

func GetRequestBody(c *gin.Context) ([]byte, error) {
	requestBody, _ := c.Get(constants.KeyRequestBody)
	if requestBody != nil {
		return requestBody.([]byte), nil
	}
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	_ = c.Request.Body.Close()
	c.Set(constants.KeyRequestBody, requestBody)
	return requestBody.([]byte), nil
}

func UnmarshalBodyReusable(c *gin.Context, v any) error {
	requestBody, err := GetRequestBody(c)
	if err != nil {
		return err
	}
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err = json.Unmarshal(requestBody, &v)
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		//这里需要重新绑定, 因为它读的c.Request的, 而c.Request需要在下面才会
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		err = BindMultipartForm(c.Request, v)
	} else {
		err = c.ShouldBind(&v)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	// Reset request body
	return nil
}

// 自定义绑定函数
func BindMultipartForm(r *http.Request, dest interface{}) error {
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to struct")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("dest must point to a struct")
	}

	// 解析 multipart 表单（注意内存限制）32M
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	// 反射获取目标结构体
	typ := val.Type()

	// 遍历结构体字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// 获取标签值（表单字段名）
		tag := field.Tag.Get("form")
		if tag == "" {
			tag = field.Name // 默认使用字段名
		}
		// 处理文件数组字段
		if field.Type.Kind() == reflect.Slice && field.Type.Elem() == reflect.TypeOf((*multipart.FileHeader)(nil)) {
			// 收集所有匹配字段名的文件
			var files []*multipart.FileHeader
			for key, headers := range r.MultipartForm.File {
				// 匹配字段名加索引的形式 (如 "image[0]")
				if strings.HasPrefix(key, tag+"[") && strings.HasSuffix(key, "]") {
					files = append(files, headers...)
				}
				// 同时匹配无索引的字段名 (如 "image")
				if key == tag {
					files = append(files, headers...)
				}
			}

			// 按索引排序文件
			sort.Slice(files, func(i, j int) bool {
				// 这里可以添加更复杂的排序逻辑
				return files[i].Filename < files[j].Filename
			})

			// 创建切片并填充
			slice := reflect.MakeSlice(field.Type, len(files), len(files))
			for idx, fh := range files {
				slice.Index(idx).Set(reflect.ValueOf(fh))
			}
			fieldVal.Set(slice)
			continue
		}

		// 处理单个文件字段（保持原逻辑）
		if field.Type == reflect.TypeOf((*multipart.FileHeader)(nil)) {
			_, header, err := r.FormFile(tag)
			if err != nil {
				// 允许文件不存在（非必填字段）
				if err != http.ErrMissingFile {
					return fmt.Errorf("field '%s' no exists : %v", tag, err)
				}
			} else {
				fieldVal.Set(reflect.ValueOf(header))
			}
			continue
		}

		// 处理文本字段
		values := r.MultipartForm.Value[tag]
		if len(values) == 0 {
			continue
		}

		// 类型转换（支持 string/int/bool 等基础类型）
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(values[0])
		case reflect.Int:
			intVal, _ := strconv.Atoi(values[0])
			fieldVal.SetInt(int64(intVal))
		case reflect.Bool:
			boolVal, _ := strconv.ParseBool(values[0])
			fieldVal.SetBool(boolVal)
		case reflect.Slice:
			// 处理文本数组字段
			if field.Type.Elem().Kind() == reflect.String {
				slice := reflect.MakeSlice(field.Type, len(values), len(values))
				for i, v := range values {
					slice.Index(i).SetString(v)
				}
				fieldVal.Set(slice)
			}
			// 可扩展其他类型...
		}
	}

	return nil
}

// Curl请求
func Curl(url string, method string, body any) (error, *http.Response) {
	jsonData, err := json.Marshal(body)
	requestBody := bytes.NewBuffer(jsonData)
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return fmt.Errorf("new request failed: %w", err), nil
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return err, nil
	}
	if resp == nil {
		return errors.New("resp is nil"), nil
	}
	_ = req.Body.Close()
	return nil, resp
}
