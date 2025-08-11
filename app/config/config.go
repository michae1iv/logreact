package config

import (
	"github.com/spf13/viper"
)

// * All configs in one struct
type Config struct {
	Server         ServerConfig      `mapstructure:"server"`       //? Настройки сервера
	Database       DatabaseConfig    `mapstructure:"database"`     //? Настройки базы данных
	Reader         ReaderConfig      `mapstructure:"reader"`       //? Настройки Reader
	Writer         WriterConfig      `mapstructure:"writer"`       //? Настройки Reader
	Logging        LoggingConfig     `mapstructure:"logging"`      //? Настройки логирования
	RuleHandler    RuleHandlerConfig `mapstructure:"rule_handler"` //? Rule Handler settings
	Authentication AuthConfig        `mapstructure:"authentication"`
}

// * ServerConfig web-iinterface configuration
type ServerConfig struct {
	Port           int      `mapstructure:"port"`            //? Порт для веб-интерфейса
	AllowedOrigins []string `mapstructure:"allowed_origins"` //? Адрес и порт на котором работает Frontend
}

// * DatabaseConfig PostgreSQL database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`     //? Адрес сервера БД
	Port     int    `mapstructure:"port"`     //? Порт для подключения
	User     string `mapstructure:"user"`     //? Имя пользователя
	Password string `mapstructure:"password"` //? Пароль пользователя
	DBName   string `mapstructure:"dbname"`   //? Название базы данных
	SSLMode  string `mapstructure:"sslmode"`  //? Режим SSL для подключения
}

// * LoggingConfig consist path to log files
type LoggingConfig struct {
	LogPath string `mapstructure:"log_path"` //? Path to folder where app log files stored
}

type RuleHandlerConfig struct {
	ReadBufferSize   int `mapstructure:"r_buff_size"` //? Size of buffer for Reader, if buffer is full Reader waiting for GlobalHandler to read message
	WriterBufferSize int `mapstructure:"w_buff_size"` //? Size of buffer for Writer, if buffer is full GlobalHandler waiting for Writer to read message
}

type AuthConfig struct {
	Secret string `mapstructure:"jwt_secret_key"`
}

// * LoadConfig loads config from file `config.yaml` and returns pointer to Config struct
func LoadConfig(path string) (*Config, error) {
	// Setting path to config and type
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// REading config from file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Parsing data in Config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ----------------------------------------------------------------------------Reader's configs----------------------------------------------------------------------------------------------------
type ReaderConfig struct {
	Kafka *KafkaConsumerConfig `mapstructure:"kafka,omitempty"`
}

// * KafkaConfig Apache Kafka configuration
type KafkaConsumerConfig struct {
	Enable          bool     `mapstructure:"enable"`
	Brokers         []string `mapstructure:"brokers"`   //? Список брокеров Kafka
	Topic           string   `mapstructure:"topic"`     //? Тема Kafka
	GroupID         string   `mapstructure:"group_id"`  //? Группа потребителей Kafka
	ClientID        string   `mapstructure:"client_id"` //? Идентификатор клиента Kafka
	AutoOffsetReset string   `mapstructure:"auto_offset_reset"`
	MaxPollRecords  int      `mapstructure:"max_poll_records"` //? Максимальное количество записей за раз
}

// ----------------------------------------------------------------------------Writer's configs----------------------------------------------------------------------------------------------------

type WriterConfig struct {
	Kafka   *KafkaProducerConfig `mapstructure:"kafka,omitempty"`
	Postgre *PostgreWriterConfig `mapstructure:"postgre,omitempty"`
}

// * KafkaConfig Apache Kafka configuration
type KafkaProducerConfig struct {
	Enable      bool     `mapstructure:"enable"`
	Brokers     []string `mapstructure:"brokers"`          //? Список брокеров Kafka
	Topic       string   `mapstructure:"topic"`            //? Тема Kafka
	Acks        string   `mapstructure:"acks"`             //? Подтверждение доставки
	Retries     int      `mapstructure:"retries"`          //? Количество попыток при неудачной отправки
	Сompression string   `mapstructure:"compression_type"` //? Тип сжатия
	LingerMS    int      `mapstructure:"linger_ms"`        //? Буфферизация
	BatchSize   int      `mapstructure:"batch_size"`       //? Размер одного батча
}

// * KafkaConfig Apache Kafka configuration
type PostgreWriterConfig struct {
	Enable bool `mapstructure:"enable"`
}
