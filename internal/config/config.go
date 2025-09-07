package config

import "os"

// getEnv obtiene una variable de entorno o devuelve el valor por defecto
func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
