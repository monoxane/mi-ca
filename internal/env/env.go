package env

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

var (
	APP_MODE             = "PROD"
	LOG_LEVEL            = "INFO"
	C2_SERVER            = "localhost:3000"
	KUBERNETES_NAMESPACE = "mi"
	CLUSTER_NAME         = ""
	CLUSTER_OWNERS       = []string{}
)

func PopulateEnvironment() bool {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			fmt.Printf("Error reading configuration file: %s\n", err)
			return false
		}
	}

	if viper.IsSet("MODE") {
		APP_MODE = viper.GetString("MODE")
		log.Printf("[ENV] Application Mode: %s", APP_MODE)
	}

	if viper.IsSet("LOG_LEVEL") {
		LOG_LEVEL = viper.GetString("LOG_LEVEL")
		log.Printf("[ENV] Log Level: %s", LOG_LEVEL)
	}

	if viper.IsSet("NAMESPACE") {
		KUBERNETES_NAMESPACE = viper.GetString("NAMESPACE")
		log.Printf("[ENV] Cluster Namespace: %s", KUBERNETES_NAMESPACE)
	} else {
		log.Print("[ENV] Cluster Namespace not set, using `mi`")
	}

	if viper.IsSet("C2_URL") {
		C2_SERVER = viper.GetString("C2_URL")
		log.Printf("[ENV] c2 server: %s", C2_SERVER)
	} else {
		log.Print("[ENV] c2 server not set")
		return false
	}

	if viper.IsSet("CLUSTER_NAME") {
		CLUSTER_NAME = viper.GetString("CLUSTER_NAME")
		log.Printf("[ENV] cluster name: %s", CLUSTER_NAME)
	} else {
		log.Print("[ENV] cluster name not set")
		return false
	}

	if viper.IsSet("CLUSTER_OWNERS") {
		CLUSTER_OWNERS = viper.GetStringSlice("CLUSTER_OWNERS")
		log.Printf("[ENV] Cluster Owners: %+v", CLUSTER_OWNERS)
	} else {
		log.Print("[ENV] Cluster Owners not set, please set it to aide in management")
	}

	return true
}
