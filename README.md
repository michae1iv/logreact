# 📜 Open-Source Streaming Security Event Correlator

## 📖 Project Description

This project implements a **real-time streaming security event correlator** with support for:

* 📌 Loading and executing **correlation rules** (JSON)
* 📌 Support for **sticky variables** for multi-step scenarios
* 📌 Data sources: **Kafka**, **HTTP API** (extensible)
* 📌 High performance (processing tens of thousands of events per second)
* 📌 Log parsing and condition checking using **regular expressions**
* 📌 Flexible architecture: easily add new data sources (`ReaderAgent`) and event processors
* 📌 Alert generation when rules are triggered

Use cases:

* 🚨 Detection of network attacks (DoS, port scanning, brute-force, etc.)
* 🛡 Log analysis for Linux, Windows, Nginx, Cisco, and other systems
* 📡 Correlation of events across different sources

---

## 📂 Project Structure

```
.
├── main.go                  # Entry point (main.go)
├── config/                  # Configuration files
├── rw/                      # ReaderAgent implementations (Kafka, HTTP)
├── rule_manager/            # Correlation logic
├── api/                     # JSON API for app
├── logger/                  # Logging
├── go.mod / go.sum          # Go dependencies
└── db                       # Models and Database driver
```

---

## ⚙️ Installation and Build

### 1️⃣ Install dependencies

Make sure you have Go **version 1.21+** installed and Kafka (if used as an event source).

```bash
git clone https://github.com/michae1iv/logreact
cd security-correlator

# Install Go dependencies
go mod tidy
```

---

### 2️⃣ Configuration

Example configuration (`config.yaml`):

```yaml
server:
  port: 8080
  allowed_origins: ["http://localhost:3000", "http://192.168.50.47:3000"]

database:
  host: "localhost"
  port: 5432
  user: "cor_admin"
  password: "cor_admin"
  dbname: "maindb"
  sslmode: "disable"

rule_handler: 
  r_buff_size: 1000
  w_buff_size: 100
  frame_buff: 100

reader:
  kafka:
    enable: true
    brokers:
      - "localhost:9092"
    topic: "logs"
    group_id: "myGroup"
    client_id: "logreact"
    auto_offset_reset: "latest"
    max_poll_records: 50000

writer:
  kafka:
    enable: true
    brokers:
      - "localhost:9092"
    topic: "alerts"
    acks: "all"               # 0 / 1 / all
    retries: 5
    compression_type: "snappy"
    linger_ms: 10
    batch_size: 16384
  postgre:
    enable: true

logging:
  log_path: "/var/log/logreact"

authentication:
  jwt_secret_key: "your_secret_key"

```

---

### 3️⃣ Build

```bash
go build -o correlator ./cmd
```

---

### 4️⃣ Run

#### From binary:

```bash
./correlator -config ./config/config.yaml
```

#### From source:

```bash
go run ./cmd -config ./config/config.yaml
```

---

## 📜 Creating Correlation Rules

Rules are described in JSON format. Example rule:

```json
{
  "rule": "SSH Brute Force",
  "ukey": "auth.log",
  "params": {
    "ttl": "1h",
    "sev_level": 3,
    "desc": "SSH Brute Force detection",
    "no_alert": false
  },
  "condition": {
    "logic": "event.program == 'sshd' AND event.message: 'Failed password' AND source.ip = $STICKY$",
    "freq": "100/min"
  },
  "alert": {
    "fields": "source.ip, event.program",
    "addfields": {
      "message": "SSH brute force attempt from %source.ip%",
      "from": "logreact",
      "key": "ssh_bruteforce"
    }
  }
}
```

📌 **Notes**:

* `ukey` — source key (log type)
* `params.ttl` — correlation container lifetime
* `condition.logic` — logical expression 
* `alert` — alert description, fields, and message text

---

## 📈 Performance

With the built-in event generator, the application can process at least **20k events/sec** with single Kafka broker.

---

## 🛠 Roadmap

* Add list of regexp
* **YARA rules** support
* Add 2FA
* Add password length check and limit on login

---

## 📜 License

This project is licensed under the **MIT** license.
