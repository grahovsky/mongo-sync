package common

import (
	"gopkg.in/ini.v1"
)

type Config struct {
	SyncDateField string
	FirstSyncDate string
	LogFile       string
	Timeout       int
	NumWorkers    int
	SrcConnStr    string
	SrcDbName     string
	SrcUsername   string
	SrcPassword   string
	SrcAuthsorce  string
	DstConnStr    string
	DstDbName     string
	DstUsername   string
	DstPassword   string
	DstAuthsorce  string
}

var config *Config

func LoadConfig(filename string) (*Config, error) {
	logger := GetLogger()

	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}

	sectionSrc := cfg.Section("source")
	sectionDst := cfg.Section("destination")
	sectionCmn := cfg.Section("common")

	config = &Config{
		SyncDateField: sectionCmn.Key("sync_date_field").String(),
		FirstSyncDate: sectionCmn.Key("first_sync_date").String(),
		LogFile:       sectionCmn.Key("log_file").String(),
		SrcConnStr:    sectionSrc.Key("src_conn_str").String(),
		SrcDbName:     sectionSrc.Key("src_db_name").String(),
		SrcUsername:   sectionSrc.Key("src_username").String(),
		SrcPassword:   sectionSrc.Key("src_password").String(),
		SrcAuthsorce:  sectionSrc.Key("src_auth_sorce").String(),
		DstConnStr:    sectionDst.Key("dst_conn_str").String(),
		DstDbName:     sectionDst.Key("dst_db_name").String(),
		DstUsername:   sectionDst.Key("dst_username").String(),
		DstPassword:   sectionDst.Key("dst_password").String(),
		DstAuthsorce:  sectionDst.Key("dst_auth_sorce").String(),
	}

	timeout, err := sectionCmn.Key("timeout").Int()
	if err != nil {
		logger.Fatal(err)
	}
	config.Timeout = timeout

	numWorkers, err := sectionCmn.Key("num_workers").Int()
	if err != nil {
		logger.Fatal(err)
	}
	config.NumWorkers = numWorkers

	return config, nil
}
