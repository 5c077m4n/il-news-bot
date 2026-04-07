terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

variable "OPENROUTER_API_KEY" {
  type      = string
  sensitive = true
}
variable "TELEGRAM_API_KEY" {
  type      = string
  sensitive = true
}
variable "TELEGRAM_BOT_ID" {
  type      = string
  sensitive = true
}
variable "MOTIS_PROVIDER" {
  type    = string
  default = "openrouter"
}
variable "MOLTIS_API_KEY" {
  type      = string
  sensitive = true
}
variable "DO_TOKEN" {
  type      = string
  sensitive = true
}

provider "digitalocean" {
  token = var.DO_TOKEN
}

variable "ssh_public_key_path" {
  type        = string
  description = "Path to the public SSH key file"
  default     = "~/.ssh/keys/do_ed25519.pub"
}
resource "digitalocean_ssh_key" "default" {
  name       = "default"
  public_key = file(var.ssh_public_key_path)
}
resource "digitalocean_droplet" "news_agents" {
  image    = "ubuntu-24-04-x64"
  name     = "il-news-bot-droplet"
  region   = "fra1"
  size     = "s-2vcpu-2gb"
  ssh_keys = [digitalocean_ssh_key.default.fingerprint]

  user_data = <<-EOF
    #!/bin/bash

    apt-get update
    apt-get install --yes git curl fail2ban ufw

    ufw allow OpenSSH
    ufw default deny incoming
    ufw --force enable
    systemctl enable fail2ban --now

    sed -i 's/^#\?PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
    systemctl restart ssh

    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sh /tmp/get-docker.sh

    git clone https://github.com/5c077m4n/il-news-bot.git /app
    cd /app

    cat <<-ENV_FILE >.env
    ENV=prod
    OPENROUTER_API_KEY=${var.OPENROUTER_API_KEY}
    TELEGRAM_API_KEY=${var.TELEGRAM_API_KEY}
    TELEGRAM_BOT_ID=${var.TELEGRAM_BOT_ID}
    MOTIS_PROVIDER=${var.MOTIS_PROVIDER}
    MOLTIS_API_KEY=${var.MOLTIS_API_KEY}
    ENV_FILE

    docker compose up -d --build --remove-orphans
  EOF
}

resource "digitalocean_firewall" "news_agents" {
  name        = "il-news-bot-firewall"
  droplet_ids = [digitalocean_droplet.news_agents.id]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}
