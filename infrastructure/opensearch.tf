# opensearch.tf

resource "awscc_opensearchserverless_collection" "kb_vector_store" {
  name = "${local.project_name}-kb-vector-store-${random_pet.suffix.id}" # Ensure globally unique name
  type = "VECTORSEARCH"
  tags = [
    { key = "Name", value = "${local.project_name}-kb-vector-store-collection" }
  ]
  # 必要に応じて standby_replicas や description を設定
}

# 暗号化ポリシー (KMSキーを使う場合など]
resource "awscc_opensearchserverless_security_policy" "encryption_policy" {
  name        = "${local.project_name}-kb-vector-encrypt-${random_pet.suffix.id}"
  type        = "encryption"
  policy      = jsonencode({
    Rules = [{ # ルールを修正
      ResourceType = "collection"
      Resource     = ["collection/${awscc_opensearchserverless_collection.kb_vector_store.name}"]
    }],
    AWSOwnedKey = true # または KMSキーを指定: KmsKeyArn = "arn:aws:kms:..."
  })
  description = "Encryption policy for the knowledge base vector store"
  provider = awscc # awscc プロバイダーを指定
}

# ネットワークポリシー (VPCアクセスの場合など]
resource "awscc_opensearchserverless_security_policy" "network_policy" {
  name        = "${local.project_name}-kb-vector-network-${random_pet.suffix.id}"
  type        = "network"
  policy      = jsonencode([{ # ルールを修正
    Rules = [{ # SourceVPCEs を削除し、AllowFromPublic を設定
      ResourceType = "collection",
      Resource     = ["collection/${awscc_opensearchserverless_collection.kb_vector_store.name}"]
    }],
    AllowFromPublic = true # VPCアクセスのみにする場合は false にし、VPCEndpointを設定
    # SourceVPCEs = [aws_vpc_endpoint.opensearch.id] # VPCエンドポイントリソースを参照する場合
  }])
  description = "Network policy for the knowledge base vector store"
  provider = awscc # awscc プロバイダーを指定
}

# データアクセスポリシー (Knowledge Base Roleからのアクセスを許可]
# 注意: このリソースは bedrock_kb_role が作成された後に適用される必要があるため、
#       明示的な依存関係が必要な場合があります。
resource "awscc_opensearchserverless_access_policy" "data_access_policy" {
  name        = "${local.project_name}-kb-vector-access-${random_pet.suffix.id}"
  type        = "data"
  policy      = jsonencode([
    {
      Rules = [
        {
          ResourceType = "collection"
          Resource     = ["collection/${awscc_opensearchserverless_collection.kb_vector_store.name}"]
          Permission   = [ # 必要な最小限の権限を確認・調整
            "aoss:DescribeCollectionItems",
            "aoss:ReadDocument"
            # Knowledge Base がインデックスを作成・管理する場合、以下の権限も必要
            # "aoss:CreateIndex",
            # "aoss:DeleteIndex",
            # "aoss:UpdateIndex",
            # "aoss:DescribeIndex",
            # "aoss:WriteDocument"
          ]
        },
        {
          ResourceType = "index"
          Resource     = ["index/${awscc_opensearchserverless_collection.kb_vector_store.name}/*"]
          Permission   = [ # 必要な最小限の権限を確認・調整
            "aoss:ReadDocument"
            # Knowledge Base がインデックスを作成・管理する場合、以下の権限も必要
            # "aoss:CreateIndex",
            # "aoss:DeleteIndex",
            # "aoss:UpdateIndex",
            # "aoss:DescribeIndex",
            # "aoss:WriteDocument"
          ]
        }
      ]
      Principal = [aws_iam_role.bedrock_kb_role.arn] # iam.tf で作成するロールを参照
    }
  ])
  description = "Data access policy for the knowledge base vector store"
  provider    = awscc # awscc プロバイダーを指定

  # bedrock_kb_role が作成された後にこのポリシーを作成するように依存関係を設定
  depends_on = [
    aws_iam_role.bedrock_kb_role
  ]
} 