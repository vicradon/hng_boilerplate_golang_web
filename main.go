package main

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/hngprojects/hng_boilerplate_golang_web/cronjobs"
	"github.com/hngprojects/hng_boilerplate_golang_web/external/request"
	"github.com/hngprojects/hng_boilerplate_golang_web/internal/config"
	"github.com/hngprojects/hng_boilerplate_golang_web/internal/models/migrations"
	"github.com/hngprojects/hng_boilerplate_golang_web/internal/models/seed"
	"github.com/hngprojects/hng_boilerplate_golang_web/pkg/repository/storage"
	"github.com/hngprojects/hng_boilerplate_golang_web/pkg/repository/storage/postgresql"
	"github.com/hngprojects/hng_boilerplate_golang_web/pkg/repository/storage/redis"
	"github.com/hngprojects/hng_boilerplate_golang_web/pkg/router"
	"github.com/hngprojects/hng_boilerplate_golang_web/utility"
)

// Initialize Prometheus metrics
var (
	cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_temperature_celsius",
		Help: "Current temperature of the CPU.",
	})
)

func init() {
	prometheus.MustRegister(cpuTemp)
}

func main() {

	cpuTemp.Set(65.3)

	logger := utility.NewLogger() //Warning !!!!! Do not recreate this action anywhere on the app
	configuration := config.Setup(logger, "./app")

	postgresql.ConnectToDatabase(logger, configuration.Database)
	redis.ConnectToRedis(logger, configuration.Redis)

	validatorRef := validator.New()

	db := storage.Connection()

	cronjobs.StartCronJob(request.ExternalRequest{Logger: logger}, *storage.DB, "send-notifications")

	if configuration.Database.Migrate {
		migrations.RunAllMigrations(db)
		// call the seed function
		seed.SeedDatabase(db.Postgresql)
	}

	r := router.Setup(logger, validatorRef, db, &configuration.App)
	r.GET("/metrics", router.PrometheusHandler())

	utility.LogAndPrint(logger, fmt.Sprintf("Server is starting at 127.0.0.1:%s", configuration.Server.Port))
	log.Fatal(r.Run(":" + configuration.Server.Port))
}
