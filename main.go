package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

//利用映射将ini配置文件解析到struct中
type mysqlConfig struct {
	Address  string `ini:"address"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}
type redisConfig struct {
	Address  string `ini:"address"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
	Test     bool   `ini:"test"`
}
type Config struct {
	mysqlConfig `ini:"mysql"`
	redisConfig `ini:"redis"`
}

//定义一个函数来解析ini配置文件内的信息
//函数接收ini文件的路径和解析到目标结构的指针类型
func loadIni(file string, cfg interface{}) error {
	var structName string
	//1.首先根据路径读取配置文件信息，将读取的信息存放到一个切片中，若路径错误则提示错误信息并退出
	fileSlice, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Open ini file failed,err:%v", err)
		return err
	}
	//2.判断用来接收解析信息的类型是否为结构体类型的指针
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr || reflect.TypeOf(cfg).Elem().Kind() != reflect.Struct {
		err = errors.New("cfg param should be a struct porint")
		return err
	}
	//3.当以上条件全部满足后，再处理获取到的配置文件信息
	//3.1.按照回车分割每一行数据
	fileString := strings.Split(string(fileSlice), "\n")
	//3.2遍历每一行数据
	for idx, value := range fileString {
		//3.3去掉字符串首尾的空格
		value = strings.TrimSpace(value)
		//3.4如果字符串是以";"或"#"开头说明是注释，则跳过,若是空行也跳过
		if strings.HasPrefix(value, ";") || strings.HasPrefix(value, "#") || len(value) == 0 {
			continue
		}
		//3.5如果字符串的首是"["尾是"]"说明该字符串是节头
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			head := fileString[idx][1 : len(value)-1]
			if len(head) == 0 {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return err
			} //t := reflect.TypeOf(data)
			for i := 0; i < reflect.TypeOf(cfg).Elem().NumField(); i++ {
				if head == reflect.TypeOf(cfg).Elem().Field(i).Tag.Get("ini") {
					structName = reflect.TypeOf(cfg).Elem().Field(i).Name
				}
			}
		} else {
			//4.如果字符串的首不是"["尾也不是"]"说明该字符串是以"="号分割的键值对
			//4.1判断字符串中是否含有"=",或"="位置不合法
			if strings.HasPrefix(value, "=") || strings.HasSuffix(value, "=") || !strings.Contains(value, "=") {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return err
			}
			//4.2以等号分割
			//返回子串str在字符串s中第一次出现的位置
			index := strings.Index(value, "=")
			key := strings.TrimSpace(value[:index])
			v := strings.TrimSpace(value[index+1:])
			//4.3根据structName将cfg里面将对应的字段给取出来
			structObjvalue := reflect.ValueOf(cfg).Elem().FieldByName(structName)
			structObjtype := structObjvalue.Type()
			//4.4 判断取出来的字段是不是一个结构体类型
			if structObjtype.Kind() != reflect.Struct {
				err = fmt.Errorf("%s should be a struct", structName)
				return err
			}
			var fileName string
			var fileType reflect.StructField
			//4.5遍历结构体的每一个字段，判断tag是不是和结构体的key相等
			for i := 0; i < structObjvalue.NumField(); i++ {
				//tag信息是存储在类型信息中
				if structObjtype.Field(i).Tag.Get("ini") == key {
					//找到对应的值
					fileName = structObjtype.Field(i).Name
					fileType = structObjtype.Field(i)
					break
				}
			}
			//4.6根据找到的fileName去取出对应的字段
			//如果fileName长度为0说明并找到相应字段，则跳过本次循环
			if len(fileName) == 0 {
				continue
			}
			fileObj := structObjvalue.FieldByName(fileName)
			switch fileType.Type.Kind() {
			case reflect.String:
				fileObj.SetString(v)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				//将字符串类型转换为10进制的int64类型
				valueInt, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return err
				}
				fileObj.SetInt(valueInt)
			case reflect.Bool:
				//将字符串类型转换为Bool类型
				valueBool, err := strconv.ParseBool(v)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return err
				}
				fileObj.SetBool(valueBool)
			case reflect.Float32, reflect.Float64:
				//将字符串类型转换为Float64类型
				valueFloat, err := strconv.ParseFloat(v, 64)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return err
				}
				fileObj.SetFloat(valueFloat)
			}

		}
	}
	return nil
}
func main() {
	var cfg Config
	err := loadIni("./mysql.ini", &cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg)
}
