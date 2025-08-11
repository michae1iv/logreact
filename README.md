# ***Logreact*** — Open-Source Streaming Security Event Correlator

## 📖 Project Description

This project implements a **real-time streaming security event correlator** with support for:

* 📌 Loading and executing **correlation rules** (JSON)
* 📌 Support for **sticky variables** for multi-step scenarios
* 📌 Data sources: **Kafka**, **PostgreSQL** (extensible)
* 📌 High performance (processing tens of thousands of events per second)
* 📌 Log parsing and condition checking using **regular expressions**
* 📌 Flexible architecture: easily add new data sources (`ReaderAgent`) and event processors
* 📌 Alert generation when rules are triggered

Use cases:

* 🚨 Detection of network attacks (DoS, port scanning, brute-force, etc.)
* 🛡 Log analysis for Linux, Windows, Nginx, Cisco, and other systems
* 📡 Correlation of events across different sources

Also there is frontend client for Logreact:

[Frontend client](https://github.com/michae1iv/logreact_front)

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

Make sure you have Go **version 1.21+** installed.

```bash
git clone https://github.com/michae1iv/logreact
cd security-correlator

# Install Go dependencies
go mod tidy
```
Install PostgreSQL, then add database, user with superuser permissions

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
go build -o correlator
```

---

### 4️⃣ Run

#### From binary:

```bash
./correlator
```

#### From source:

```bash
go run main.go
```

---

## 📜 Creating Correlation Rules

Rules are described in JSON format.

Operations:
|Operation|Priority|Meaning|
|:-|:-:|:-:|
|AND / and|1|logical multiplication|
|OR / or|0|logical sum|
|:|2|checks if field contains value on right side|
|= / ==|2|checks if field has same value on right side|
|-> / contains|2|checks if field's value in list|
|!:|2|NOT :|
|!=|2|NOT =|
|!->|2|NOT ->|

Example rule:

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

## 🛠 Add Data Source

**To add data source you need:** 
1. Open file config/config.go and find Reader or Writer agent config then add new source, like this:
```
type WriterConfig struct {
	Kafka   *KafkaProducerConfig `mapstructure:"kafka,omitempty"` // same name
	Postgre *PostgreWriterConfig `mapstructure:"postgre,omitempty"` // same name
}

// * KafkaConfig Apache Kafka configuration
type KafkaProducerConfig struct {
	Enable      bool     `mapstructure:"enable"`
	Brokers     []string `mapstructure:"brokers"`
	Topic       string   `mapstructure:"topic"`
	Acks        string   `mapstructure:"acks"`
	Retries     int      `mapstructure:"retries"`
	Сompression string   `mapstructure:"compression_type"`
	LingerMS    int      `mapstructure:"linger_ms"`
	BatchSize   int      `mapstructure:"batch_size"`
}

```
3. Open file rw/[reader.go/writer.go], look at methods thats need to be defined
4. Add package with your new agent
5. Define new agent in reader.go **NOTE: name of agent in config.go and [reader.go/writer.go] should be the same**:
```
// Struct for defining agents and their struct, if you want to include new reader you need to write a struct and methods thats defined in ReaderAgent
type Writers struct {
	Kafka   *kafka.KafkaW // There must be a channel for each writer agent
	Postgre *postgre.PostgreW
}

// Initialising new readers here
var writers = &Writers{
	Kafka:   kafka.NewWriter(), // same name
	Postgre: postgre.NewWriter(), // same name
}
```

---

## 📈 Performance

With the built-in event generator, the application can process at least **20k events/sec** with single Kafka broker.

Stand configuration:
* OS: Ubuntu Ubuntu 24.04.2 LTS
* SSD: 256 Gb
* Memory: 32 Gb
* Processor: AMD Ryzen 7 5700x
* Single Apache Kafka Broker
* Single Logreact app

---

## 🛠 Roadmap

* Add list of regexp
* **YARA rules** support
* Add 2FA
* Add password length check and limit on login

---

## 📜 License

This project is licensed under the **MIT** license.
