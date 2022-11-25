package types

import (
	"encoding/json"
	"github.com/apache/iotdb-client-go/client"
	"github.com/apache/iotdb-client-go/rpc"
	"iotdb-courier/emitter/internal/config"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

type RecordSet struct {
	DeviceIds         []string
	MeasurementsSlice [][]string
	DataTypes         [][]client.TSDataType
	Values            [][]interface{}
	Timestamps        []int64
	mutex             sync.Mutex
}

func (r *RecordSet) AddRecord(message *Message, config *config.Config) error {
	var msgVal map[string]interface{}
	d := json.NewDecoder(strings.NewReader(string(message.Value)))
	d.UseNumber()
	if err := d.Decode(&msgVal); err != nil {
		return err
	}
	deviceId := determineDeviceId(message, msgVal, config)
	var measurementsSlice []string
	var dataTypes []client.TSDataType
	var values []interface{}
	var timestamps int64
	for key, value := range msgVal {
		if key == config.IoTDB.TimeField {
			// 处理时间字段
			timeStr := msgVal[key].(string)
			location, _ := time.LoadLocation(config.IoTDB.TimeZone)
			parsed, err := time.ParseInLocation(config.IoTDB.TimeParse, timeStr, location)
			if err != nil {
				return err
			}
			if parsed.Year() == 0 {
				parsed = parsed.AddDate(time.Now().Year(), 0, 0)
			}
			if config.IoTDB.TimeUnit == "ns" {
				timestamps = parsed.UnixNano()
			} else if config.IoTDB.TimeUnit == "ms" {
				timestamps = parsed.UnixMilli()
			}
		} else if reflect.TypeOf(value).String() == "string" {
			// 处理字符串序列
			measurementsSlice = append(measurementsSlice, key)
			dataTypes = append(dataTypes, client.TEXT)
			values = append(values, value)
		} else if reflect.TypeOf(value).String() == "json.Number" {
			// 处理数值序列
			measurementsSlice = append(measurementsSlice, key)
			dataTypes = append(dataTypes, client.DOUBLE)
			f, _ := value.(json.Number).Float64()
			values = append(values, f)
		}
	}
	r.DeviceIds = append(r.DeviceIds, deviceId)
	r.MeasurementsSlice = append(r.MeasurementsSlice, measurementsSlice)
	r.DataTypes = append(r.DataTypes, dataTypes)
	r.Values = append(r.Values, values)
	r.Timestamps = append(r.Timestamps, timestamps)
	return nil
}

func determineDeviceId(message *Message, msgVal map[string]interface{}, config *config.Config) string {
	deviceId := config.IoTDB.BasePath
	if len(config.IoTDB.OverwriteDeviceId) > 0 {
		deviceId = deviceId + "." + config.IoTDB.OverwriteDeviceId
		reg := regexp.MustCompile("#\\{([^}]+)\\}")
		found := reg.FindAllStringSubmatch(deviceId, -1)
		for _, sub := range found {
			if val, ok := msgVal[sub[1]]; ok {
				deviceId = strings.Replace(deviceId, sub[0], val.(string), -1)
			} else {
				deviceId = strings.Replace(deviceId, sub[0], "", -1)
			}
		}
	} else {
		if len(message.Tag) > 0 {
			deviceId += "." + message.Tag
		} else {
			idx := strings.Index(message.Key, ":")
			if idx > 0 {
				deviceId += "." + message.Key[0:idx]
			} else {
				deviceId += "." + message.Key
			}
		}
	}
	return deviceId
}

func (r *RecordSet) Len() int {
	return len(r.Timestamps)
}

func (r *RecordSet) Reset() {
	r.DeviceIds = r.DeviceIds[0:0]
	r.MeasurementsSlice = r.MeasurementsSlice[0:0]
	r.DataTypes = r.DataTypes[0:0]
	r.Values = r.Values[0:0]
	r.Timestamps = r.Timestamps[0:0]
}

func (r *RecordSet) Insert(session *client.Session) (*rpc.TSStatus, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.Len() == 0 {
		return nil, nil
	}
	tsStatus, err := session.InsertAlignedRecords(r.DeviceIds, r.MeasurementsSlice, r.DataTypes, r.Values, r.Timestamps)
	if err != nil {
		return nil, err
	}
	r.Reset()
	return tsStatus, nil
}
