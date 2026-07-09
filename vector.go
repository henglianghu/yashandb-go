package yasdb

/*
#include "yacapi.h"
*/
import "C"

import (
	"database/sql/driver"
	"fmt"
	"unsafe"
)

// VectorFormat 向量格式枚举
type VectorFormat uint8

const (
	VectorFormatFlex    VectorFormat = 0 // 灵活格式
	VectorFormatFloat16 VectorFormat = 1 // 半精度浮点
	VectorFormatFloat32 VectorFormat = 2 // 单精度浮点
	VectorFormatFloat64 VectorFormat = 3 // 双精度浮点
	VectorFormatInt8    VectorFormat = 4 // 8位整数
)

// Vector 向量数据类型
type Vector struct {
	Data   interface{}  // []float32, []float64, 或 []int8
	Dim    uint16       // 向量维度
	Format VectorFormat // 格式类型
}

// Value 实现 driver.Valuer 接口。
// 返回 Vector 的字符串表示，供 gorm 等上层 ORM 做日志格式化。
// 实际的参数绑定由 YasStmt.CheckNamedValue 直接识别 Vector 类型走专用路径，
// 不依赖 Value() 的返回值。注意：不能返回 v 自身，否则 gorm logger 会无限递归。
func (v Vector) Value() (driver.Value, error) {
	if v.Data == nil {
		return nil, nil
	}
	return v.String(), nil
}

// String 返回 Vector 的字符串表示
func (v Vector) String() string {
	switch data := v.Data.(type) {
	case []float32:
		result := "["
		for i, f := range data {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%g", f)
		}
		result += "]"
		return result
	case []float64:
		result := "["
		for i, f := range data {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%g", f)
		}
		result += "]"
		return result
	case []int8:
		result := "["
		for i, n := range data {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%d", n)
		}
		result += "]"
		return result
	case string:
		return data
	default:
		return ""
	}
}

// Scan 实现 sql.Scanner 接口，支持从数据库扫描 Vector 数据
func (v *Vector) Scan(src interface{}) error {
	if src == nil {
		*v = Vector{}
		return nil
	}
	switch s := src.(type) {
	case *Vector:
		*v = *s
		return nil
	case Vector:
		*v = s
		return nil
	default:
		return &VectorScanError{Type: typeName(src)}
	}
}

type VectorScanError struct {
	Type string
}

func (e *VectorScanError) Error() string {
	return "unsupported Vector Scan type: " + e.Type
}

func typeName(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case *Vector:
		return "*yasdb.Vector"
	case Vector:
		return "yasdb.Vector"
	default:
		return "unknown"
	}
}

// createVectorBufferByFormat 根据 VectorFormat 创建对应的缓冲区
func createVectorBufferByFormat(format VectorFormat, dim uint16) (interface{}, uint32, error) {
	switch format {
	case VectorFormatFloat32:
		buf := make([]float32, dim)
		return buf, uint32(dim * 4), nil
	case VectorFormatFloat64:
		buf := make([]float64, dim)
		return buf, uint32(dim * 8), nil
	case VectorFormatInt8:
		buf := make([]int8, dim)
		return buf, uint32(dim), nil
	case VectorFormatFlex:
		// Flex 格式默认作为 FLOAT32 处理
		buf := make([]float32, dim)
		return buf, uint32(dim * 4), nil
	default:
		return nil, 0, fmt.Errorf("unsupported vector format: %v", format)
	}
}

// getVectorElementSizeByFormat 返回格式对应的元素大小
func getVectorElementSizeByFormat(format VectorFormat) int {
	switch format {
	case VectorFormatFloat32:
		return 4
	case VectorFormatFloat64:
		return 8
	case VectorFormatInt8:
		return 1
	case VectorFormatFlex:
		return 4 // Flex 默认为 float32
	default:
		return 0
	}
}

// vectorBindInfo 包含 Vector 绑定所需的信息
type vectorBindInfo struct {
	desc      unsafe.Pointer // descriptor 指针
	yacType   C.YapiType
	bindSize  C.int32_t
	bufLength C.int32_t
	indicator *C.int32_t
	value     C.YapiPointer
	freeType  valueFreeType
}

// bindVector 创建 Vector 的绑定信息
// env: Yapi 环境指针
// v: Vector 值
// 返回绑定信息和错误
func bindVector(env *C.YapiEnv, v Vector) (*vectorBindInfo, error) {
	desc := new(unsafe.Pointer)
	if err := yapiDescAlloc2(env, desc, C.YAPI_DESC_VECTOR); err != nil {
		return nil, err
	}
	vector := (*C.YapiVector)(*desc)

	// 根据数据创建 vector
	var err error
	switch data := v.Data.(type) {
	case []float32:
		elemSize := getVectorElementSizeByFormat(VectorFormatFloat32)
		dim := C.uint16_t(len(data))
		array := (*C.uint8_t)(unsafe.Pointer(&data[0]))
		arrayLen := C.uint32_t(len(data) * elemSize)
		err = yapiVectorFromArray(vector, C.YAPI_VECTOR_FORMAT_FLOAT32, dim, array, arrayLen, 0)
	case []float64:
		elemSize := getVectorElementSizeByFormat(VectorFormatFloat64)
		dim := C.uint16_t(len(data))
		array := (*C.uint8_t)(unsafe.Pointer(&data[0]))
		arrayLen := C.uint32_t(len(data) * elemSize)
		err = yapiVectorFromArray(vector, C.YAPI_VECTOR_FORMAT_FLOAT64, dim, array, arrayLen, 0)
	case []int8:
		dim := C.uint16_t(len(data))
		array := (*C.uint8_t)(unsafe.Pointer(&data[0]))
		arrayLen := C.uint32_t(len(data))
		err = yapiVectorFromArray(vector, C.YAPI_VECTOR_FORMAT_INT8, dim, array, arrayLen, 0)
	case string:
		text := C.CString(data)
		defer C.free(unsafe.Pointer(text))
		textLen := C.uint32_t(len(data))
		err = yapiVectorFromText(vector, C.YAPI_VECTOR_FORMAT_FLEX, 0, text, textLen, 0)
	default:
		yapiDescFree2(env, *desc, C.YAPI_TYPE_VECTOR)
		return nil, fmt.Errorf("unsupported Vector data type: %T", v.Data)
	}

	if err != nil {
		yapiDescFree2(env, *desc, C.YAPI_TYPE_VECTOR)
		return nil, err
	}

	return &vectorBindInfo{
		desc:      unsafe.Pointer(desc),
		yacType:   C.YAPI_TYPE_VECTOR,
		bindSize:  -1,
		bufLength: -1,
		indicator: nil,
		value:     C.YapiPointer(unsafe.Pointer(desc)),
		freeType:  vectorFree,
	}, nil
}

// bindVectorPointer 创建 *Vector 的绑定信息
func bindVectorPointer(env *C.YapiEnv, v *Vector) (*vectorBindInfo, error) {
	if v == nil {
		return nil, fmt.Errorf("nil Vector pointer")
	}
	return bindVector(env, *v)
}

// bindVectorSlice 创建 []Vector 的绑定信息（使用第一个 Vector 作为模板）
func bindVectorSlice(env *C.YapiEnv, vectors []Vector) (*vectorBindInfo, error) {
	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty Vector slice")
	}
	return bindVector(env, vectors[0])
}

// bindPointerVectorSlice 创建 []*Vector 的绑定信息（使用第一个 Vector 作为模板）
func bindPointerVectorSlice(env *C.YapiEnv, vectors []*Vector) (*vectorBindInfo, error) {
	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty Vector pointer slice")
	}
	if vectors[0] == nil {
		return nil, fmt.Errorf("nil Vector in slice")
	}
	return bindVector(env, *vectors[0])
}

// bindVectorValue 统一的 Vector 绑定入口，支持 Vector、*Vector、[]Vector、[]*Vector
func bindVectorValue(env *C.YapiEnv, v interface{}) (*vectorBindInfo, error) {
	switch val := v.(type) {
	case Vector:
		return bindVector(env, val)
	case *Vector:
		return bindVectorPointer(env, val)
	case []Vector:
		return bindVectorSlice(env, val)
	case []*Vector:
		return bindPointerVectorSlice(env, val)
	default:
		return nil, fmt.Errorf("unsupported Vector type: %T", v)
	}
}
