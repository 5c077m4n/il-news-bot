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
variable "ssh_allowed_ip" {
  type        = string
  description = "Your public IP, allowed to SSH to the instance (no /32 suffix)"
  sensitive   = true
}
variable "ssh_public_key_path" {
  type    = string
  default = "~/.ssh/keys/aws_ed25519.pub"
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
  description = "Firewall for il-news-bot"

  ingress {
    description = "SSH"
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["${var.ssh_allowed_ip}/32"]
  }

  egress {
    description      = "All outbound"
    protocol         = "-1"
    from_port        = 0
    to_port          = 0
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
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
        ]
      },
      {
        Effect   = "Allow"
        Action   = ["kms:Decrypt"]
        Resource = [data.aws_kms_alias.ssm.target_key_arn]
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

resource "aws_key_pair" "news_agents" {
  key_name   = "il-news-bot"
  public_key = file(pathexpand(var.ssh_public_key_path))
}

resource "aws_launch_template" "news_agents" {
  name          = "il-news-bot"
  image_id      = data.aws_ssm_parameter.al2023_ami.value
  instance_type = "t3.small"
  key_name      = aws_key_pair.news_agents.key_name

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
      encrypted = true
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
      mkdir -p /opt/il-news-bot
      chown ec2-user:ec2-user /opt/il-news-bot

      cat >/etc/ssh/sshd_config.d/60-hardening.conf <<'EOF'
      PasswordAuthentication no
      KbdInteractiveAuthentication no
      PermitRootLogin no
      EOF

      chmod 600 /etc/ssh/sshd_config.d/60-hardening.conf
      sshd -t && systemctl restart sshd

      dnf install -y dnf-automatic
      sed -i -E 's/^#?[[:space:]]*upgrade_type.*/upgrade_type = security/; s/^#?[[:space:]]*apply_updates.*/apply_updates = yes/' /etc/dnf/automatic.conf
      systemctl enable --now dnf-automatic.timer

      if ! command -v fail2ban-client >/dev/null 2>&1; then
        dnf install -y python3-pip python3-systemd
        pip3 install --upgrade fail2ban
      fi
      cat >/etc/fail2ban/jail.d/sshd.local <<'EOF'
      [sshd]
      enabled  = true
      backend  = systemd
      port     = ssh
      maxretry = 5
      findtime = 10m
      bantime  = 1h
      EOF
      cat >/etc/systemd/system/fail2ban.service <<'EOF'
      [Unit]
      Description=Fail2Ban Service
      After=network.target sshd.service

      [Service]
      Type=forking
      ExecStart=/usr/local/bin/fail2ban-server -xf start
      ExecStop=/usr/local/bin/fail2ban-client stop
      PIDFile=/run/fail2ban/fail2ban.pid
      Restart=on-failure

      [Install]
      WantedBy=multi-user.target
      EOF

      systemctl daemon-reload
      systemctl enable --now fail2ban

      cat >/usr/local/bin/il-news-bot-env.sh <<'EOF'
      #!/bin/bash

      set -euo pipefail

      if ! command -v aws >/dev/null 2>&1; then
        dnf install -y aws-cli
      fi

      mkdir -p /opt/il-news-bot
      umask 077

      env_file="/opt/il-news-bot/.env"
      printf 'ENV=prod\n' >> "$env_file"
      for name in OPENROUTER_API_KEY TELEGRAM_BOT_TOKEN TELEGRAM_API_ID TELEGRAM_API_HASH; do
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

data "aws_instances" "news_agents" {
  filter {
    name   = "tag:aws:autoscaling:groupName"
    values = [aws_autoscaling_group.news_agents.name]
  }

  filter {
    name   = "instance-state-name"
    values = ["running"]
  }

  depends_on = [aws_autoscaling_group.news_agents]
}

resource "null_resource" "deploy" {
  triggers = {
    deploy_id = local.deploy_id
  }

  connection {
    type        = "ssh"
    host        = data.aws_instances.news_agents.public_ips[0]
    user        = "ec2-user"
    private_key = file(trimsuffix(pathexpand(var.ssh_public_key_path), ".pub"))
  }

  provisioner "local-exec" {
    command = "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=\"-w -s\" -a -installsuffix cgo -o main ."
  }
  provisioner "remote-exec" {
    inline = ["sudo cloud-init status --wait"]
  }
  provisioner "file" {
    source      = "${path.module}/main"
    destination = "/opt/il-news-bot/main.new"
  }
  provisioner "file" {
    source      = "${path.module}/telegram_session.data"
    destination = "/opt/il-news-bot/telegram_session.data"
  }
  provisioner "remote-exec" {
    inline = [
      "sudo mv --force /opt/il-news-bot/main.new /opt/il-news-bot/main",
      "sudo chmod 755 /opt/il-news-bot/main",
      "sudo systemctl enable il-news-bot",
      "sudo systemctl restart il-news-bot",
    ]
  }

  depends_on = [aws_autoscaling_group.news_agents]
}
