// customized error type

package __

type ErrorResp struct {
	StatusCode    StatusCode `json:"code"`
	StatusMessage string     `json:"message"`
}

func (e *ErrorResp) Error() string {
	return e.StatusMessage
}
