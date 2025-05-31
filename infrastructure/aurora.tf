# Aurora Serverless v2 + pgvector for Bedrock Knowledge Base

# ランダムなパスワード生成
resource "random_password" "aurora_password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# DB サブネットグループ
resource "aws_db_subnet_group" "aurora" {
  name       = "${local.project_name}-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-db-subnet-group"
    }
  )
}

# Aurora Serverless v2 クラスターパラメータグループ
resource "aws_rds_cluster_parameter_group" "aurora_pg15" {
  name        = "${local.project_name}-aurora-pg15-params"
  family      = "aurora-postgresql15"
  description = "Aurora PostgreSQL 15 parameter group with pgvector extension"

  # pgvector 拡張のための設定
  parameter {
    name  = "shared_preload_libraries"
    value = "pg_stat_statements,pgvector"
    apply_method = "pending-reboot"
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-aurora-pg15-params"
    }
  )
}

# Aurora Serverless v2 クラスター
resource "aws_rds_cluster" "aurora" {
  cluster_identifier      = "${local.project_name}-kb-vector-store"
  engine                  = "aurora-postgresql"
  engine_version          = "15.4"
  availability_zones      = [for az in ["a", "c"] : "${var.aws_region}${az}"]
  database_name           = "vectordb"
  master_username         = "pgadmin"
  master_password         = random_password.aurora_password.result
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  db_subnet_group_name    = aws_db_subnet_group.aurora.name
  vpc_security_group_ids  = [aws_security_group.aurora.id]
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.aurora_pg15.name
  
  # Serverless v2 設定
  engine_mode            = "provisioned"
  serverlessv2_scaling_configuration {
    min_capacity = 0.5  # 最小 ACU (0.5 = ほぼ停止状態)
    max_capacity = 2.0  # 最大 ACU (サンプルアプリなので小さめ)
  }

  # 自動スケールダウン設定（7日間アイドル後に停止）
  scaling_configuration {
    auto_pause               = true
    max_capacity             = 2
    min_capacity             = 0
    seconds_until_auto_pause = 300  # 5分後に停止
    timeout_action           = "ForceApplyCapacityChange"
  }

  # その他の設定
  enabled_cloudwatch_logs_exports = ["postgresql"]
  storage_encrypted               = true
  
  # クリーンアップ設定
  skip_final_snapshot     = true
  deletion_protection     = false
  apply_immediately       = true

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-kb-vector-store"
    }
  )

  # VPC と依存関係
  depends_on = [
    aws_vpc.main,
    aws_subnet.private
  ]
}

# Aurora インスタンス
resource "aws_rds_cluster_instance" "aurora_instances" {
  count                = 1  # サンプルアプリなので1インスタンスのみ
  identifier           = "${local.project_name}-kb-vector-store-${count.index}"
  cluster_identifier   = aws_rds_cluster.aurora.id
  instance_class       = "db.serverless"
  engine               = aws_rds_cluster.aurora.engine
  engine_version       = aws_rds_cluster.aurora.engine_version
  db_subnet_group_name = aws_db_subnet_group.aurora.name
  
  # その他の設定
  publicly_accessible  = false
  
  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-kb-vector-store-${count.index}"
    }
  )
}

# Aurora 用セキュリティグループ
resource "aws_security_group" "aurora" {
  name        = "${local.project_name}-aurora-sg"
  description = "Security group for Aurora PostgreSQL with pgvector"
  vpc_id      = aws_vpc.main.id

  # PostgreSQL アクセス許可（VPC内から）
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
    description = "PostgreSQL access from within VPC"
  }

  # 全ての送信トラフィックを許可
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-aurora-sg"
    }
  )
}

# Secrets Manager に認証情報を保存
resource "aws_secretsmanager_secret" "aurora_credentials" {
  name        = "${local.project_name}/aurora-credentials"
  description = "Aurora PostgreSQL credentials for Bedrock Knowledge Base"
  
  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-aurora-credentials"
    }
  )
}

# Secrets Manager に保存する値
resource "aws_secretsmanager_secret_version" "aurora_credentials" {
  secret_id = aws_secretsmanager_secret.aurora_credentials.id
  secret_string = jsonencode({
    username = aws_rds_cluster.aurora.master_username
    password = aws_rds_cluster.aurora.master_password
    engine   = "postgres"
    host     = aws_rds_cluster.aurora.endpoint
    port     = 5432
    dbname   = aws_rds_cluster.aurora.database_name
  })
}

# pgvector 拡張機能を有効化するための初期化スクリプト
# NOTE: 実際の初期化は AWS CLI または PostgreSQL クライアントで手動実行が必要
output "pgvector_init_script" {
  value = <<-EOF
    -- ログイン方法:
    -- PGPASSWORD='${random_password.aurora_password.result}' psql -h ${aws_rds_cluster.aurora.endpoint} -U ${aws_rds_cluster.aurora.master_username} -d ${aws_rds_cluster.aurora.database_name}
    
    -- pgvector 拡張機能の有効化
    CREATE EXTENSION IF NOT EXISTS vector;
    
    -- Knowledge Base 用テーブル作成
    CREATE TABLE knowledge_chunks (
      id UUID PRIMARY KEY,
      embedding VECTOR(1536),
      text TEXT,
      metadata JSONB
    );
    
    -- HNSW インデックス作成（検索高速化）
    CREATE INDEX ON knowledge_chunks
    USING hnsw (embedding vector_l2_ops)
    WITH (m = 16, ef_construction = 64);
  EOF
  sensitive = true
} 