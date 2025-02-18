package models

type SignVo struct {
	ClientId     *int64     `json:"clientId"`
	Timestamp    string     `json:"timestamp"`
	RandomString string     `json:"randomString"`
	ExpiresTime  CustomTime `json:"expiresTime"`
	Host         string     `json:"host"`
	Sign         string     `json:"sign"`
}

// 自定义 UnmarshalJSON 方法处理时间戳
// func (s *SignVo) UnmarshalJSON(data []byte) error {

// 	// 1 定义别名：
// 	// 这里定义了一个新的类型 AliasSignVo，它是 SignVo 的别名。虽然它们的底层类型相同，但 Go 会将它们视为不同的类型。这意味着 AliasSignVo 不会继承 SignVo 的方法（包括自定义的 UnmarshalJSON）。
// 	type AliasSignVo SignVo

// 	// 2 创建一个匿名结构体：

// 	// 这里定义了一个匿名结构体，它包含两个字段：
// 	// ExpiresTime int64：用于接收 JSON 中的 expiresTime 字段。
// 	// *AliasSignVo：这是一个匿名字段，它嵌入了 AliasSignVo 类型，允许我们在匿名结构体中直接访问 SignVo 的其他字段。
// 	// 通过这种方式，我们可以将 SignVo 的其他字段委托给 AliasSignVo ，同时单独处理 expiresTime 字段。
// 	aux := &struct {
// 		ExpiresTime int64 `json:"expiresTime"`
// 		*AliasSignVo
// 	}{
// 		AliasSignVo: (*AliasSignVo)(s),
// 	}

// 	if err := json.Unmarshal(data, aux); err != nil {
// 		return err
// 	}

// 	// 将时间转换戳为 time.Time
// 	s.ExpiresTime = time.UnixMilli(aux.ExpiresTime)
// 	return nil
// }
