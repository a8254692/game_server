package stack

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

func deepCopy(dst, src reflect.Value, fordb bool) {
	switch src.Kind() {
	case reflect.Interface:
		value := src.Elem()
		if !value.IsValid() {
			return
		}
		newValue := reflect.New(value.Type()).Elem()
		deepCopy(newValue, value, fordb)
		dst.Set(newValue)
	case reflect.Ptr:
		value := src.Elem()
		if !value.IsValid() {
			return
		}
		dst.Set(reflect.New(value.Type()))
		deepCopy(dst.Elem(), value, fordb)
	case reflect.Map:
		dst.Set(reflect.MakeMap(src.Type()))
		keys := src.MapKeys()
		for _, key := range keys {
			value := src.MapIndex(key)
			newValue := reflect.New(value.Type()).Elem()
			deepCopy(newValue, value, fordb)
			dst.SetMapIndex(key, newValue)
		}
	case reflect.Slice:
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			deepCopy(dst.Index(i), src.Index(i), fordb)
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			value := src.Field(i)
			//time类型特殊处理
			if src.Field(i).Type() == reflect.TypeOf(time.Time{}) {
				newValue := reflect.New(src.Field(i).Type())
				srData, _ := value.Interface().(time.Time).MarshalBinary()
				newValue.Interface().(*time.Time).UnmarshalBinary(srData)
				dst.Field(i).Set(newValue.Elem())
				continue
			}

			if src.Type().Field(i).Name == "XXX_unrecognized" {
				continue
			}

			if fordb && src.Type().Field(i).Tag.Get("bson") == "-" {
				continue
			}

			if value.CanSet() {
				deepCopy(dst.Field(i), value, fordb)
			}
		}
	default:
		dst.Set(src)
	}
}

// 异步存数据库专用，会跳过bson"-"标签的字段，普通拷贝不可以使用
func DeepCloneForDB(v interface{}) interface{} {
	dst := reflect.New(reflect.TypeOf(v)).Elem()
	deepCopy(dst, reflect.ValueOf(v), true)
	return dst.Interface()
}

func DeepClone(v interface{}) interface{} {
	dst := reflect.New(reflect.TypeOf(v)).Elem()
	deepCopy(dst, reflect.ValueOf(v), false)
	return dst.Interface()
}

func DeepCopy(dst, src interface{}) {
	typeDst := reflect.TypeOf(dst)
	typeSrc := reflect.TypeOf(src)
	if typeDst != typeSrc {
		panic("DeepCopy: " + typeDst.String() + " != " + typeSrc.String())
	}
	if typeSrc.Kind() != reflect.Ptr {
		panic("DeepCopy: pass arguments by address")
	}

	valueDst := reflect.ValueOf(dst).Elem()
	valueSrc := reflect.ValueOf(src).Elem()
	if !valueDst.IsValid() || !valueSrc.IsValid() {
		panic("DeepCopy: invalid arguments")
	}

	deepCopy(valueDst, valueSrc, false)
}

// 用b的所有字段覆盖a的
// 如果fields不为空, 表示用b的特定字段覆盖a的
// a应该为结构体指针
func CopyFields(a interface{}, b interface{}, fields ...string) bool {
	at := reflect.TypeOf(a)
	av := reflect.ValueOf(a)
	bt := reflect.TypeOf(b)
	bv := reflect.ValueOf(b)
	// 简单判断下
	if at.Kind() != reflect.Ptr {
		panic("-->unitcom CopyFields:a must be a struct pointer")
		return false
	}
	av = reflect.ValueOf(av.Interface())
	// 要复制哪些字段
	_fields := make([]string, 0)
	if len(fields) > 0 {
		_fields = fields
	} else {
		for i := 0; i < bv.NumField(); i++ {
			_fields = append(_fields, bt.Field(i).Name)
		}
	}
	if len(_fields) == 0 {
		panic("-->unitcom CopyFields:no fields to copy")
		return false
	}
	// 复制
	for i := 0; i < len(_fields); i++ {
		name := _fields[i]
		f := av.Elem().FieldByName(name)
		bValue := bv.FieldByName(name)
		// a中有同名的字段并且类型一致才复制
		if f.IsValid() && f.Kind() == bValue.Kind() {
			f.Set(bValue)
		}
		// else {
		// 	panic("-->unitcom CopyFields:no such field or different kind")
		// 	//return false
		// }
	}
	return true
}

// 复制src的同名同类型属性值给dst
// dst应该为结构体指针 src可以是结构体，也可以是结构体指针
func SimpleCopyProperties(dst, src interface{}) (err error) {
	// 防止意外panic
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%v", e))
		}
	}()

	dstType, dstValue := reflect.TypeOf(dst), reflect.ValueOf(dst)
	srcType, srcValue := reflect.TypeOf(src), reflect.ValueOf(src)

	// dst必须结构体指针类型
	if dstType.Kind() != reflect.Ptr || dstType.Elem().Kind() != reflect.Struct {
		return errors.New("dst type should be a struct pointer")
	}

	// src必须为结构体或者结构体指针，.Elem()类似于*ptr的操作返回指针指向的地址反射类型
	if srcType.Kind() == reflect.Ptr {
		srcType, srcValue = srcType.Elem(), srcValue.Elem()
	}
	if srcType.Kind() != reflect.Struct {
		return errors.New("src type should be a struct or a struct pointer")
	}

	// 取具体内容
	dstType, dstValue = dstType.Elem(), dstValue.Elem()

	// 属性个数
	propertyNums := dstType.NumField()

	for i := 0; i < propertyNums; i++ {
		// 属性
		property := dstType.Field(i)
		// 待填充属性值
		propertyValue := srcValue.FieldByName(property.Name)

		// 无效，说明src没有这个属性 || 属性同名但类型不同
		if !propertyValue.IsValid() || property.Type != propertyValue.Type() {
			continue
		}

		if dstValue.Field(i).CanSet() {
			dstValue.Field(i).Set(propertyValue)
		}
	}

	return nil
}

// struct转换为[]byte   binary包处理二进制
func StructToBytes_Binary(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.LittleEndian, data)
	// if err != nil {
	// 	panic(err)
	// }
	return buf.Bytes(), err
}

// 将[]byte转换为struct   binary包处理二进制
func BytesToStruct_Binary(b []byte, data interface{}) error {
	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.LittleEndian, data)
	return err
}

// struct转为[]byte
func StructToBytes_Unsafe(iter interface{}, len int) []byte {
	var x reflect.SliceHeader
	x.Len = len
	x.Cap = len
	x.Data = reflect.ValueOf(iter).Pointer()
	return *(*[]byte)(unsafe.Pointer(&x))

	// var sizeOfMyStruct = int(unsafe.Sizeof(entity.EntityAcc{}))
	// buf := stack.StructToBytes_Unsafe(tStruct, sizeOfMyStruct)
}

// []byte转为struct
func BytesToStruct_Unsafe(buf []byte) unsafe.Pointer {
	return unsafe.Pointer(
		(*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data,
	)

	//tStruct := (*tStruct)(stack.BytesToStruct_Unsafe(buf))
}

// struct转为[]byte
func StructToBytes_Gob(iter interface{}) ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(iter)
	if err != nil {
		//fmt.Println(err)
	}
	return b.Bytes(), err
}

// []byte转为struct
func BytesToStruct_Gob(buf []byte, iter interface{}) error {
	b := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(b)
	err := dec.Decode(iter)
	if err != nil {
		//fmt.Println("Error decoding GOB data:", err)
	}
	return err
}

// 不同结构之间相同字段赋值
func StructCopySame_Json(dst, src interface{}) error {
	data, errS := json.Marshal(&src)
	if errS != nil {
		return errS
	}
	errD := json.Unmarshal(data, &dst)
	if errD != nil {
		return errD
	}
	return nil
}

// 不同结构之间相同字段赋值 没调通
func StructCopySame_Gob(dst, src interface{}) error {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	errS := enc.Encode(src)
	if errS != nil {
		//return errS
	}
	dec := gob.NewDecoder(&b)
	errD := dec.Decode(dst)
	if errD != nil {
		//return errD
	}
	return nil
}
