terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

variable "aws_profile" {
  type    = string
  default = "default"
}
variable "aws_region" {
  type    = string
  default = "il-central-1"
}

variable "OPENROUTER_API_KEY" {
  type      = string
  sensitive = true
}
variable "TELEGRAM_BOT_TOKEN" {
  type      = string
  sensitive = true
}
variable "TELEGRAM_API_ID" {
  type      = string
  sensitive = true
}
variable "TELEGRAM_API_HASH" {
  type      = string
  sensitive = true
}
variable "TELEGRAM_PHONE_NUMBER" {
  type      = string
  sensitive = true
}
provider "aws" {
  profile = var.aws_profile
  region  = var.aws_region
}

locals {
  deploy_id = formatdate("YYYYMMDDhhmmss", timestamp())
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

resource "aws_security_group" "news_agents" {
  name        = "il-news-bot-firewall"
  description = "Firewall for il-news-bot (no ingress, HTTPS-only egress)"

  egress {
    description      = "HTTPS outbound"
    protocol         = "tcp"
    from_port        = 443
    to_port          = 443
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "artifacts" {
  bucket        = "il-news-bot-artifacts-${data.aws_caller_identity.current.account_id}"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "artifacts" {
  bucket                  = aws_s3_bucket.artifacts.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  versioning_configuration {
    status = "Enabled"
  }
}

data "aws_iam_policy_document" "artifacts_kms" {
  statement {
    sid       = "AllowAccountFullControl"
    effect    = "Allow"
    actions   = ["kms:*"]
    resources = ["*"]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

  statement {
    sid       = "AllowVPCFlowLogDelivery"
    effect    = "Allow"
    actions   = ["kms:GenerateDataKey*"]
    resources = ["*"]
    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    sid    = "AllowAutoScalingSLRCreateGrants"
    effect = "Allow"
    actions = [
      "kms:CreateGrant",
      "kms:ListGrants",
      "kms:RevokeGrant",
      "kms:GenerateDataKeyWithoutPlaintext",
      "kms:DescribeKey",
    ]
    resources = ["*"]
    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling",
      ]
    }
    condition {
      test     = "StringLike"
      variable = "aws:PrincipalArn"
      values   = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"]
    }
  }
}

resource "aws_kms_key" "artifacts" {
  description             = "il-news-bot artifacts encryption"
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.artifacts_kms.json
}

resource "aws_s3_bucket_server_side_encryption_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    bucket_key_enabled = true

    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.artifacts.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    id     = "expire-old-deploy-versions"
    status = "Enabled"
    filter { prefix = "deploy/" }
    noncurrent_version_expiration { noncurrent_days = 3 }
    abort_incomplete_multipart_upload { days_after_initiation = 1 }
  }

  rule {
    id     = "expire-flow-logs"
    status = "Enabled"
    filter { prefix = "flow-logs/" }
    expiration { days = 14 }
  }
}

data "aws_iam_policy_document" "artifacts" {
  statement {
    sid     = "DenyInsecureTransport"
    effect  = "Deny"
    actions = ["s3:*"]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [aws_s3_bucket.artifacts.arn, "${aws_s3_bucket.artifacts.arn}/*"]
    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
    condition {
      test     = "ArnNotLike"
      variable = "aws:SourceArn"
      values   = ["arn:aws:logs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:*"]
    }
  }

  statement {
    sid     = "DenyNonAccountPrincipals"
    effect  = "Deny"
    actions = ["s3:*"]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [aws_s3_bucket.artifacts.arn, "${aws_s3_bucket.artifacts.arn}/*"]
    condition {
      test     = "StringNotEquals"
      variable = "aws:PrincipalAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
    condition {
      test     = "ArnNotLike"
      variable = "aws:SourceArn"
      values   = ["arn:aws:logs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:*"]
    }
  }

  statement {
    sid     = "AllowFlowLogAclCheck"
    effect  = "Allow"
    actions = ["s3:GetBucketAcl"]
    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
    resources = [aws_s3_bucket.artifacts.arn]
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["arn:aws:logs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:*"]
    }
  }

  statement {
    sid     = "DenyNonKMSUploads"
    effect  = "Deny"
    actions = ["s3:PutObject"]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = ["${aws_s3_bucket.artifacts.arn}/*"]
    condition {
      test     = "StringEquals"
      variable = "s3:x-amz-server-side-encryption"
      values   = ["AES256"]
    }
  }

  statement {
    sid     = "AllowVPCFlowLogsDelivery"
    effect  = "Allow"
    actions = ["s3:PutObject"]
    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
    resources = [
      "${aws_s3_bucket.artifacts.arn}/flow-logs/AWSLogs/${data.aws_caller_identity.current.account_id}/*",
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["arn:aws:logs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:*"]
    }
  }
}

resource "aws_flow_log" "default_vpc" {
  vpc_id                   = data.aws_vpc.default.id
  traffic_type             = "ALL"
  max_aggregation_interval = 60
  log_destination_type     = "s3"
  log_destination          = "${aws_s3_bucket.artifacts.arn}/flow-logs"
}

resource "aws_s3_bucket_policy" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  policy = data.aws_iam_policy_document.artifacts.json
}

resource "aws_ssm_parameter" "openrouter_api_key" {
  name  = "/il-news-bot/OPENROUTER_API_KEY"
  type  = "SecureString"
  value = var.OPENROUTER_API_KEY
}
resource "aws_ssm_parameter" "telegram_bot_token" {
  name  = "/il-news-bot/TELEGRAM_BOT_TOKEN"
  type  = "SecureString"
  value = var.TELEGRAM_BOT_TOKEN
}
resource "aws_ssm_parameter" "telegram_api_id" {
  name  = "/il-news-bot/TELEGRAM_API_ID"
  type  = "SecureString"
  value = var.TELEGRAM_API_ID
}
resource "aws_ssm_parameter" "telegram_api_hash" {
  name  = "/il-news-bot/TELEGRAM_API_HASH"
  type  = "SecureString"
  value = var.TELEGRAM_API_HASH
}
resource "aws_ssm_parameter" "telegram_phone_number" {
  name  = "/il-news-bot/TELEGRAM_PHONE_NUMBER"
  type  = "SecureString"
  value = var.TELEGRAM_PHONE_NUMBER
}

data "aws_iam_policy_document" "ec2_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

data "aws_kms_alias" "ssm" {
  name = "alias/aws/ssm"
}

resource "aws_iam_role" "ec2_instance" {
  name               = "il-news-bot-ec2-instance"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume.json
}

resource "aws_iam_role_policy" "ec2_instance_ssm" {
  name = "ssm-parameters"
  role = aws_iam_role.ec2_instance.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = ["ssm:GetParameter", "ssm:GetParameters"]
        Resource = [
          aws_ssm_parameter.openrouter_api_key.arn,
          aws_ssm_parameter.telegram_bot_token.arn,
          aws_ssm_parameter.telegram_api_id.arn,
          aws_ssm_parameter.telegram_api_hash.arn,
          aws_ssm_parameter.telegram_phone_number.arn,
        ]
      },
      {
        Effect   = "Allow"
        Action   = ["kms:Decrypt"]
        Resource = [data.aws_kms_alias.ssm.target_key_arn]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:CreateGrant",
          "kms:DescribeKey",
          "kms:GenerateDataKey*",
          "kms:ReEncrypt*",
        ]
        Resource = [aws_kms_key.artifacts.arn]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ec2_instance_ssm_core" {
  role       = aws_iam_role.ec2_instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role_policy" "ec2_instance_s3" {
  name = "deploy-artifacts"
  role = aws_iam_role.ec2_instance.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject"]
        Resource = ["${aws_s3_bucket.artifacts.arn}/deploy/*"]
      }
    ]
  })
}
resource "aws_iam_instance_profile" "ec2_instance" {
  name = "il-news-bot-ec2-instance"
  role = aws_iam_role.ec2_instance.name
}

data "aws_ssm_parameter" "al2023_ami" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
}

resource "aws_launch_template" "news_agents" {
  name          = "il-news-bot"
  image_id      = data.aws_ssm_parameter.al2023_ami.value
  instance_type = "t3.small"

  iam_instance_profile {
    name = aws_iam_instance_profile.ec2_instance.name
  }

  network_interfaces {
    associate_public_ip_address = true
    security_groups             = [aws_security_group.news_agents.id]
  }

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      volume_size           = 20
      volume_type           = "gp3"
      encrypted             = true
      kms_key_id            = aws_kms_key.artifacts.arn
      delete_on_termination = true
    }
  }

  metadata_options {
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
    instance_metadata_tags      = "disabled"
  }

  user_data = base64encode(
    <<-EOT
      #!/bin/bash

      set -euo pipefail

      mkdir -p /opt/il-news-bot
      chown ec2-user:ec2-user /opt/il-news-bot

      systemctl disable --now sshd.service

      dnf install -y dnf-automatic
      sed -i -E 's/^#?[[:space:]]*upgrade_type.*/upgrade_type = security/; s/^#?[[:space:]]*apply_updates.*/apply_updates = yes/' /etc/dnf/automatic.conf
      systemctl enable --now dnf-automatic.timer

      cat >/usr/local/bin/il-news-bot-env.sh <<'EOF'
      #!/bin/bash

      set -euo pipefail

      if ! command -v aws >/dev/null 2>&1; then
        dnf install -y aws-cli
      fi

      mkdir -p /opt/il-news-bot
      umask 077

      env_file="/opt/il-news-bot/.env"
      printf 'ENV=prod\n' > "$env_file"
      for name in OPENROUTER_API_KEY TELEGRAM_BOT_TOKEN TELEGRAM_API_ID TELEGRAM_API_HASH TELEGRAM_PHONE_NUMBER; do
        value=""
        for i in $(seq 1 12); do
          if value=$(aws ssm get-parameter --name "/il-news-bot/$name" --with-decryption --query Parameter.Value --output text --region ${var.aws_region} 2>/dev/null); then
            break
          fi
          sleep 10
        done
        if [ -z "$value" ]; then
          echo "failed to fetch /il-news-bot/$name from SSM after retries" >&2
          exit 1
        fi
        printf '%s=%s\n' "$name" "$value" >> "$env_file"
      done
      chown ec2-user:ec2-user /opt/il-news-bot/.env
      EOF
      chmod 700 /usr/local/bin/il-news-bot-env.sh

      cat >/etc/systemd/system/il-news-bot.service <<'EOF'
      [Unit]
      Description=il-news-bot
      After=network-online.target
      Wants=network-online.target

      [Service]
      Type=simple
      User=ec2-user
      WorkingDirectory=/opt/il-news-bot
      ExecStartPre=+/usr/local/bin/il-news-bot-env.sh
      ExecStart=/opt/il-news-bot/main
      Restart=on-failure
      RestartSec=5

      [Install]
      WantedBy=multi-user.target
      EOF

      systemctl daemon-reload
      systemctl enable il-news-bot
    EOT
  )
}

resource "aws_autoscaling_group" "news_agents" {
  name                = "il-news-bot-ec2-${aws_launch_template.news_agents.latest_version}"
  vpc_zone_identifier = data.aws_subnets.default.ids
  min_size            = 1
  max_size            = 1
  desired_capacity    = 1
  health_check_type   = "EC2"

  launch_template {
    id      = aws_launch_template.news_agents.id
    version = aws_launch_template.news_agents.latest_version
  }
}

resource "null_resource" "deploy" {
  triggers = {
    deploy_id = local.deploy_id
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -euo pipefail

      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -a -installsuffix cgo -o main .
      aws s3 cp --region ${var.aws_region} main s3://${aws_s3_bucket.artifacts.bucket}/deploy/main
      aws s3 cp --region ${var.aws_region} telegram_session.data s3://${aws_s3_bucket.artifacts.bucket}/deploy/telegram_session.data
    EOT
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -euo pipefail

      instance_id=""
      for i in $(seq 1 60); do
        instance_id=$(aws ec2 describe-instances --region ${var.aws_region} --filters "Name=tag:aws:autoscaling:groupName,Values=${aws_autoscaling_group.news_agents.name}" "Name=instance-state-name,Values=running" --query "Reservations[].Instances[].InstanceId" --output text)
        if [ -n "$instance_id" ] && [ "$instance_id" != "None" ]; then
          break
        fi
        echo "waiting for a running instance (attempt $i/60)..."
        sleep 10
      done
      if [ -z "$instance_id" ] || [ "$instance_id" = "None" ]; then
        echo "no running instance found" >&2
        exit 1
      fi

      command_id=""
      for i in $(seq 1 60); do
        if command_id=$(aws ssm send-command \
          --region ${var.aws_region} \
          --instance-ids "$instance_id" \
          --document-name "AWS-RunShellScript" \
          --comment "il-news-bot deploy" \
          --parameters '{
            "commands": [
              "cloud-init status --wait",
              "command -v aws >/dev/null 2>&1 || dnf install -y -q aws-cli",
              "mkdir -p /opt/il-news-bot",
              "aws s3 cp --region ${var.aws_region} s3://${aws_s3_bucket.artifacts.bucket}/deploy/main /opt/il-news-bot/main.new",
              "aws s3 cp --region ${var.aws_region} s3://${aws_s3_bucket.artifacts.bucket}/deploy/telegram_session.data /opt/il-news-bot/telegram_session.data",
              "mv --force /opt/il-news-bot/main.new /opt/il-news-bot/main",
              "chmod 755 /opt/il-news-bot/main",
              "chown ec2-user:ec2-user /opt/il-news-bot/main /opt/il-news-bot/telegram_session.data",
              "chmod 600 /opt/il-news-bot/telegram_session.data",
              "systemctl enable il-news-bot",
              "systemctl restart il-news-bot"
            ]
          }' \
          --query "Command.CommandId" \
          --output text 2>&1); then
          break
        fi
        echo "waiting for the SSM agent (attempt $i/60): $command_id" >&2
        command_id=""
        sleep 10
      done
      if [ -z "$command_id" ]; then
        echo "failed to send the deploy command via SSM" >&2
        exit 1
      fi

      aws ssm wait command-executed --region ${var.aws_region} --command-id "$command_id" --instance-id "$instance_id" 2>/dev/null || true
      status=$(aws ssm get-command-invocation --region ${var.aws_region} --command-id "$command_id" --instance-id "$instance_id" --query "Status" --output text)
      aws ssm get-command-invocation --region ${var.aws_region} --command-id "$command_id" --instance-id "$instance_id" --query "[StandardOutputContent,StandardErrorContent]" --output text
      if [ "$status" != "Success" ]; then
        echo "deploy command finished with status: $status" >&2
        exit 1
      fi
    EOT
  }

  depends_on = [aws_autoscaling_group.news_agents, aws_s3_bucket.artifacts]
}
