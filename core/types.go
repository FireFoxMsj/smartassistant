package core

type CallService struct {
	Domain      string
	ID          int
	Service     string
	ServiceData M      `json:"service_data"`
	Type        string // CallService
}

func (cs *CallService) Reset(clientKey string) {
	cs.Domain = ""
	cs.ID = 0
	cs.Service = ""
	cs.Type = ""

	cs.ServiceData = make(M)
	cs.ServiceData["client_key"] = clientKey
}

type CallResponse struct {
	ID      int                    `json:"id"`
	Error   string                 `json:"error,omitempty"`
	Result  map[string]interface{} `json:"result"`
	Success bool                   `json:"success"`
}

func (cr *CallResponse) AddResult(key string, value interface{}) {
	if cr.Result == nil {
		cr.Result = make(map[string]interface{})
	}
	cr.Result[key] = value
}

type serviceFunc func(args M) error

type M map[string]interface{}

func (sd M) Get(key string) interface{} {
	return sd[key]
}

func (sd M) ValInt(key string) int {
	if v := sd.Get(key); v == nil {
		return 0
	} else {
		if i, ok := v.(float64); ok {
			return int(i)
		}
		return v.(int)
	}
}
func (sd M) ValString(key string) string {
	if v := sd.Get(key); v == nil {
		return ""
	} else {
		return v.(string)
	}
}

func (sd M) Put(key string, v interface{}) {
	sd[key] = v
}
