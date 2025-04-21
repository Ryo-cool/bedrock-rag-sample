# --- S3 Bucket for Document Uploads & Knowledge Base --- #
resource "aws_s3_bucket" "documents" {
  bucket = "${local.project_name}-documents-${random_pet.suffix.id}" # Ensure globally unique bucket name

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-documents-bucket"
    }
  )
}

# Block public access to the bucket
resource "aws_s3_bucket_public_access_block" "documents_block" {
  bucket = aws_s3_bucket.documents.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Optional: Enable versioning for the bucket
resource "aws_s3_bucket_versioning" "documents_versioning" {
  bucket = aws_s3_bucket.documents.id
  versioning_configuration {
    status = "Enabled"
  }
}

# Optional: Bucket lifecycle configuration (e.g., move old files to Glacier)
# resource "aws_s3_bucket_lifecycle_configuration" "documents_lifecycle" {
#   bucket = aws_s3_bucket.documents.id
#
#   rule {
#     id     = "log"
#     status = "Enabled"
#
#     transition {
#       days          = 30
#       storage_class = "STANDARD_IA"
#     }
#
#     transition {
#       days          = 60
#       storage_class = "GLACIER"
#     }
#
#     expiration {
#       days = 90
#     }
#   }
# }

# --- S3 Bucket for Textract Results (Optional) --- #
# If you plan to store Textract output separately
# resource "aws_s3_bucket" "textract_results" {
#   bucket = "${local.project_name}-textract-results-${random_pet.suffix.id}"
#
#   tags = merge(
#     local.tags,
#     {
#       Name = "${local.project_name}-textract-results-bucket"
#     }
#   )
# }
#
# resource "aws_s3_bucket_public_access_block" "textract_results_block" {
#   bucket = aws_s3_bucket.textract_results.id
#   block_public_acls = true
#   block_public_policy = true
#   ignore_public_acls = true
#   restrict_public_buckets = true
# } 