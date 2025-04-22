// src/types/api.ts

// バックエンド API のリクエスト/レスポンス型定義
// 注意: これらは仮定義であり、実際のバックエンド仕様に合わせて変更が必要です。

// --- /summarize/text --- 
export type SummarizeTextRequest = {
  text: string;
  // 必要に応じて、要約の長さなどのパラメータを追加
};

export type SummarizeTextResponse = {
  summary: string;
};

// --- /summarize/file --- 
export type SummarizeFileRequest = {
  file_id: string; // アップロードされたファイルIDなど
  // 必要に応じて、要約の長さなどのパラメータを追加
};

export type SummarizeFileResponse = {
  summary: string;
};

// --- /qa --- 
export type Message = {
  role: 'user' | 'assistant';
  content: string;
};

export type QaRequest = {
  messages: Message[]; // チャット履歴を含むメッセージ配列
  // 必要に応じて、knowledge_base_id などのパラメータを追加
};

export type QaResponse = {
  answer: string; // アシスタントの回答
  // 必要に応じて、参照ソースなどの情報を追加
  sources?: { source_url: string; snippet: string }[];
};

// --- /upload --- 
// リクエストは FormData なので、ここではレスポンスのみ定義
export type UploadFileResponse = {
  file_id: string;
  filename: string;
  message: string;
};

// --- /document/process --- 
export type ProcessDocumentRequest = {
  file_id: string;
};

export type ProcessDocumentResponse = {
  status: 'processing' | 'completed' | 'failed';
  message?: string;
};

// --- /recommend --- 
export type RecommendRequest = {
  query?: string; // QAの質問など
  document_id?: string; // 関連文書のIDなど
};

export type DocumentSnippet = {
  document_id: string;
  title?: string; // ドキュメントのタイトル（あれば）
  snippet: string; // 関連性の高い部分のスニペット
  score?: number; // 類似度スコア（あれば）
}

export type RecommendResponse = {
  recommendations: DocumentSnippet[];
}; 