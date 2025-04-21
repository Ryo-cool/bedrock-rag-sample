package models

// ErrorDetail はAPIエラーレスポンスの詳細部分を表す構造体
type ErrorDetail struct {
	Code    string `json:"code"`              // エラーコード (例: INVALID_INPUT)
	Message string `json:"message"`           // ユーザーフレンドリーなメッセージ
	Details string `json:"details,omitempty"` // 詳細情報 (省略可能)
}

// ErrorResponse は標準的なAPIエラーレスポンスの構造体
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
