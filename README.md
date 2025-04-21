# Bedrock RAG Sample Project

## 概要 (Overview)

AWS Bedrock の RAG 機能（Knowledge Base + Bedrock Agents）を活用し、外部ドキュメントや画像から情報を検索・要約・生成するサンプルアプリケーションです。

This is a sample application that utilizes AWS Bedrock's RAG capabilities (Knowledge Base + Bedrock Agents) to search, summarize, and generate information from external documents and images.

## 主な機能 (Features)

- **自由テキスト要約 (Free Text Summarization):** 入力されたテキストを要約します。
- **PDF/画像からの要約 (Summarization from PDF/Image):** PDF や画像ファイルをアップロードし、OCR (Amazon Textract) を利用してテキストを抽出し、要約します。
- **RAG Q&A チャット (RAG Q&A Chat):** ドキュメントストアに対して自然言語で質問応答を行います。
- **類似文書レコメンド (Similar Document Recommendation):** アップロードされたドキュメントに基づいて類似する文書を推薦します。

## アーキテクチャ (Architecture)

- **フロントエンド (Frontend):** Next.js (React) - Vercel にデプロイ
- **バックエンド (Backend):** Golang BFF (Echo/Gin) - AWS Fargate (API Gateway 経由)
- **ストレージ (Storage):** Amazon S3 (原文、解析結果保存)
- **ドキュメント処理 (Document Processing):** Amazon Textract (OCR)
- **RAG / ベクターストア (RAG / Vector Store):** Bedrock Knowledge Base (S3/DynamoDB 連携)
- **レコメンド (Recommendation):** Embedding ベース (DynamoDB/OpenSearch)
- **IaC:** Terraform または AWS SDK
- **CI/CD:** Vercel (Frontend), GitHub Actions/CodePipeline (Backend/IaC)

## MVP 機能 (MVP Features - MoSCoW)

- **Must:** 自由テキスト要約, RAG Q&A チャット
- **Should:** PDF/画像要約
- **Could:** 類似文書レコメンド
- **Won't:** 高度なダッシュボード/分析

## 開発フェーズ (Development Phases)

1.  **環境セットアップ (Environment Setup):** Terraform によるインフラ構築、Vercel/Fargate 接続確認。
2.  **アップロード & S3 登録 (Upload & S3 Registration):** ファイルアップロード機能と S3 への保存。
3.  **テキスト要約 & RAG Q&A (Text Summarization & RAG Q&A - Must):** コア機能の実装。
4.  **PDF/画像要約 (PDF/Image Summarization - Should):** Textract 連携。
5.  **レコメンド機能 (Recommendation - Could):** 類似文書検索機能の実装。
6.  **テスト & E2E 検証 (Testing & E2E Verification):** 全体テスト。

## セットアップ (Setup)

(今後追記 - To be added)

```bash
# 例: Clone the repository
git clone <repository-url>
cd bedrock-rag-sample

# Install frontend dependencies
cd frontend
npm install

# Install backend dependencies (if needed)
cd ../backend
go mod download

# Configure AWS credentials
# ...

# Deploy infrastructure (Terraform)
cd ../infrastructure
terraform init
terraform apply

# Run frontend (development)
cd ../frontend
npm run dev

# Run backend (development)
cd ../backend
go run main.go
```

## 次のアクション (Next Actions)

- 「自由テキスト要約 + RAG Q&A」から着手し、コアバリューを早期に検証。
- 必要な AWS Bedrock や Textract の権限設定、簡易プロトタイプの設計を進める。
