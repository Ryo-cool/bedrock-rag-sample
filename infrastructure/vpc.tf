# --- VPC --- #
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-vpc"
    }
  )
}

# --- Subnets --- #
data "aws_availability_zones" "available" {}

resource "aws_subnet" "public" {
  count             = 2 # Create 2 public subnets in different AZs
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true # Enable auto-assign public IP for instances in public subnets

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-public-subnet-${count.index + 1}"
    }
  )
}

resource "aws_subnet" "private" {
  count             = 2 # Create 2 private subnets in different AZs
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index + 2) # Offset CIDR block index
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-private-subnet-${count.index + 1}"
    }
  )
}

# --- Internet Gateway --- #
resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-igw"
    }
  )
}

# --- NAT Gateway --- #
resource "aws_eip" "nat" {
  domain     = "vpc" # Updated from 'vpc = true' for newer AWS provider versions
  depends_on = [aws_internet_gateway.gw]

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-nat-eip"
    }
  )
}

resource "aws_nat_gateway" "gw" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id # Place NAT GW in the first public subnet

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-nat-gw"
    }
  )

  # Ensure Internet Gateway is created before NAT Gateway
  depends_on = [aws_internet_gateway.gw]
}

# --- Route Tables --- #
# Public Route Table
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-public-rt"
    }
  )
}

resource "aws_route_table_association" "public" {
  count          = length(aws_subnet.public)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Private Route Table
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.gw.id
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-private-rt"
    }
  )
}

resource "aws_route_table_association" "private" {
  count          = length(aws_subnet.private)
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
} 