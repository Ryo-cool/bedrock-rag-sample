package dto

// ErrorDetail はAPIエラーレスポンスの詳細を表す
type ErrorDetail struct {
	Code    string `json:"code"`              // エラーコード (例: "INVALID_PARAMETER")
	Message string `json:"message"`           // ユーザー向けエラーメッセージ
	Details string `json:"details,omitempty"` // (オプション) 詳細なエラー情報
}

// ErrorResponse はAPIエラーレスポンスの全体構造を表す
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
